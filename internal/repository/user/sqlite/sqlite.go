package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"whitelist/internal/core"
	domainUser "whitelist/internal/domain/user"
)

type IQueryable interface {
	Begin() (*sql.Tx, error)
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type SQLiteUserRepository struct {
	db IQueryable
}

func NewSQLiteUserRepository(db IQueryable) *SQLiteUserRepository {
	return &SQLiteUserRepository{db: db}
}

func (r *SQLiteUserRepository) UserByTelegramID(ctx context.Context, telegramID int64) (domainUser.User, error) {
	q := New(r.db)

	dbUser, err := q.UserByTelegramID(ctx, domainUser.TelegramID(telegramID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainUser.User{}, core.ErrUserNotFound
		}
		return domainUser.User{}, fmt.Errorf("failed to get user by telegram ID: %w", err)
	}

	createdAt, err := time.Parse("2006-01-02T15:04:05-0700", dbUser.CreatedAt)
	if err != nil {
		return domainUser.User{}, fmt.Errorf("failed to parse createdAt: %w", err)
	}
	updatedAt, err := time.Parse("2006-01-02T15:04:05-0700", dbUser.UpdatedAt)
	if err != nil {
		return domainUser.User{}, fmt.Errorf("failed to parse updatedAt: %w", err)
	}

	user, err := domainUser.NewUserBuilder().
		IDFromString(dbUser.ID).
		TelegramID(int64(dbUser.TelegramID)).
		FirstName(string(*dbUser.FirstName)).
		LastName(string(*dbUser.LastName)).
		Username(string(*dbUser.Username)).
		CreatedAt(createdAt).
		UpdatedAt(updatedAt).
		Build()

	if err != nil {
		return domainUser.User{}, fmt.Errorf("failed to build user: %w", err)
	}
	return user, nil
}

func (r *SQLiteUserRepository) CreateUser(ctx context.Context, user domainUser.User) error {
	q := New(r.db)

	firstName := user.FirstName()
	lastName := user.LastName()
	username := user.Username()

	err := q.CreateUser(ctx, CreateUserParams{
		ID:         user.ID().String(),
		TelegramID: user.TelegramID(),
		FirstName:  &firstName,
		LastName:   &lastName,
		Username:   &username,
		CreatedAt:  user.CreatedAt().Format("2006-01-02T15:04:05-0700"),
		UpdatedAt:  user.UpdatedAt().Format("2006-01-02T15:04:05-0700"),
	})
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *SQLiteUserRepository) UpdateUser(ctx context.Context, user domainUser.User) error {
	q := New(r.db)

	firstName := user.FirstName()
	lastName := user.LastName()
	username := user.Username()

	err := q.UpdateUser(ctx, UpdateUserParams{
		ID:         user.ID().String(),
		TelegramID: user.TelegramID(),
		FirstName:  &firstName,
		LastName:   &lastName,
		Username:   &username,
		UpdatedAt:  user.UpdatedAt().Format("2006-01-02T15:04:05-0700"),
	})
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}
