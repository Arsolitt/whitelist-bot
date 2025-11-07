package main

import (
	"database/sql"
	"log/slog"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type RequestStatus string

const (
	StatusPending  RequestStatus = "pending"
	StatusApproved RequestStatus = "approved"
	StatusRejected RequestStatus = "rejected"
)

type WhitelistRequest struct {
	ID        int64
	UserID    int64
	Username  string
	Nickname  string
	Status    RequestStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Database struct {
	db *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	slog.Info("Connecting to database", "path", dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		slog.Error("Failed to open database", "error", err, "path", dbPath)
		return nil, err
	}

	if err := db.Ping(); err != nil {
		slog.Error("Failed to ping database", "error", err)
		return nil, err
	}

	slog.Info("Database connection established successfully")

	database := &Database{db: db}
	if err := database.createTables(); err != nil {
		slog.Error("Failed to create tables", "error", err)
		return nil, err
	}

	slog.Info("Database tables initialized")
	return database, nil
}

func (d *Database) createTables() error {
	slog.Debug("Creating database tables if not exist")

	query := `
	CREATE TABLE IF NOT EXISTS whitelist_requests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		username TEXT,
		nickname TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_user_id ON whitelist_requests(user_id);
	CREATE INDEX IF NOT EXISTS idx_status ON whitelist_requests(status);
	`

	_, err := d.db.Exec(query)
	if err != nil {
		slog.Error("Failed to create tables", "error", err)
	} else {
		slog.Debug("Tables created or already exist")
	}
	return err
}

func (d *Database) CreateRequest(userID int64, username, nickname string) error {
	slog.Info("Creating new whitelist request",
		"user_id", userID,
		"username", username,
		"nickname", nickname)

	query := `
		INSERT INTO whitelist_requests (user_id, username, nickname, status)
		VALUES (?, ?, ?, ?)
	`
	result, err := d.db.Exec(query, userID, username, nickname, StatusPending)
	if err != nil {
		slog.Error("Failed to create request",
			"error", err,
			"user_id", userID,
			"nickname", nickname)
		return err
	}

	requestID, _ := result.LastInsertId()
	slog.Info("Whitelist request created successfully",
		"request_id", requestID,
		"user_id", userID,
		"nickname", nickname)

	return nil
}

func (d *Database) GetPendingRequests() ([]WhitelistRequest, error) {
	slog.Info("Fetching all pending requests")

	query := `
		SELECT id, user_id, username, nickname, status, created_at, updated_at
		FROM whitelist_requests
		WHERE status = ?
		ORDER BY created_at DESC
	`

	rows, err := d.db.Query(query, StatusPending)
	if err != nil {
		slog.Error("Failed to fetch pending requests", "error", err)
		return nil, err
	}
	defer rows.Close()

	var requests []WhitelistRequest
	for rows.Next() {
		var r WhitelistRequest
		var username sql.NullString
		err := rows.Scan(&r.ID, &r.UserID, &username, &r.Nickname, &r.Status, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			slog.Error("Failed to scan request row", "error", err)
			return nil, err
		}
		if username.Valid {
			r.Username = username.String
		}
		requests = append(requests, r)
	}

	slog.Info("Pending requests fetched successfully", "count", len(requests))
	return requests, rows.Err()
}

func (d *Database) GetRequestByID(requestID int64) (*WhitelistRequest, error) {
	slog.Debug("Fetching request by ID", "request_id", requestID)

	query := `
		SELECT id, user_id, username, nickname, status, created_at, updated_at
		FROM whitelist_requests
		WHERE id = ?
	`

	var r WhitelistRequest
	var username sql.NullString
	err := d.db.QueryRow(query, requestID).Scan(
		&r.ID, &r.UserID, &username, &r.Nickname, &r.Status, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		slog.Error("Failed to fetch request by ID",
			"error", err,
			"request_id", requestID)
		return nil, err
	}
	if username.Valid {
		r.Username = username.String
	}

	slog.Debug("Request fetched successfully",
		"request_id", requestID,
		"user_id", r.UserID,
		"status", r.Status)

	return &r, nil
}

func (d *Database) UpdateRequestStatus(requestID int64, status RequestStatus) error {
	slog.Info("Updating request status",
		"request_id", requestID,
		"new_status", status)

	query := `
		UPDATE whitelist_requests
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	result, err := d.db.Exec(query, status, requestID)
	if err != nil {
		slog.Error("Failed to update request status",
			"error", err,
			"request_id", requestID,
			"status", status)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	slog.Info("Request status updated successfully",
		"request_id", requestID,
		"status", status,
		"rows_affected", rowsAffected)

	return nil
}

func (d *Database) GetUserLastRequest(userID int64) (*WhitelistRequest, error) {
	slog.Debug("Fetching last request for user", "user_id", userID)

	query := `
		SELECT id, user_id, username, nickname, status, created_at, updated_at
		FROM whitelist_requests
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	var r WhitelistRequest
	var username sql.NullString
	err := d.db.QueryRow(query, userID).Scan(
		&r.ID, &r.UserID, &username, &r.Nickname, &r.Status, &r.CreatedAt, &r.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		slog.Debug("No requests found for user", "user_id", userID)
		return nil, nil
	}
	if err != nil {
		slog.Error("Failed to fetch user last request",
			"error", err,
			"user_id", userID)
		return nil, err
	}
	if username.Valid {
		r.Username = username.String
	}

	slog.Debug("Last request fetched successfully",
		"user_id", userID,
		"request_id", r.ID,
		"status", r.Status)

	return &r, nil
}

func (d *Database) Close() error {
	slog.Info("Closing database connection")
	err := d.db.Close()
	if err != nil {
		slog.Error("Failed to close database", "error", err)
	} else {
		slog.Info("Database connection closed successfully")
	}
	return err
}
