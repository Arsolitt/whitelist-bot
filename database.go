package main

import (
	"database/sql"
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
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	database := &Database{db: db}
	if err := database.createTables(); err != nil {
		return nil, err
	}

	return database, nil
}

func (d *Database) createTables() error {
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
	return err
}

func (d *Database) CreateRequest(userID int64, username, nickname string) error {
	query := `
		INSERT INTO whitelist_requests (user_id, username, nickname, status)
		VALUES (?, ?, ?, ?)
	`
	_, err := d.db.Exec(query, userID, username, nickname, StatusPending)
	return err
}

func (d *Database) GetPendingRequests() ([]WhitelistRequest, error) {
	query := `
		SELECT id, user_id, username, nickname, status, created_at, updated_at
		FROM whitelist_requests
		WHERE status = ?
		ORDER BY created_at DESC
	`

	rows, err := d.db.Query(query, StatusPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []WhitelistRequest
	for rows.Next() {
		var r WhitelistRequest
		var username sql.NullString
		err := rows.Scan(&r.ID, &r.UserID, &username, &r.Nickname, &r.Status, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if username.Valid {
			r.Username = username.String
		}
		requests = append(requests, r)
	}

	return requests, rows.Err()
}

func (d *Database) GetRequestByID(requestID int64) (*WhitelistRequest, error) {
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
		return nil, err
	}
	if username.Valid {
		r.Username = username.String
	}

	return &r, nil
}

func (d *Database) UpdateRequestStatus(requestID int64, status RequestStatus) error {
	query := `
		UPDATE whitelist_requests
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := d.db.Exec(query, status, requestID)
	return err
}

func (d *Database) GetUserLastRequest(userID int64) (*WhitelistRequest, error) {
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
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if username.Valid {
		r.Username = username.String
	}

	return &r, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}
