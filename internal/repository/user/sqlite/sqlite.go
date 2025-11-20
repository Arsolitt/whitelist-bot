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

const SQLITE_TIME_FORMAT = "2006-01-02T15:04:05-0700"

type iQueryable interface {
	Begin() (*sql.Tx, error)
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type UserRepository struct {
	db iQueryable
}

func NewUserRepository(db iQueryable) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) UserByTelegramID(ctx context.Context, telegramID int64) (domainUser.User, error) {
	q := New(r.db)

	dbUser, err := q.UserByTelegramID(ctx, domainUser.TelegramID(telegramID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainUser.User{}, core.ErrUserNotFound
		}
		return domainUser.User{}, fmt.Errorf("failed to get user by telegram ID: %w", err)
	}

	createdAt, err := time.Parse(SQLITE_TIME_FORMAT, dbUser.CreatedAt)
	if err != nil {
		return domainUser.User{}, fmt.Errorf("failed to parse createdAt: %w", err)
	}
	updatedAt, err := time.Parse(SQLITE_TIME_FORMAT, dbUser.UpdatedAt)
	if err != nil {
		return domainUser.User{}, fmt.Errorf("failed to parse updatedAt: %w", err)
	}

	user, err := domainUser.NewBuilder().
		IDFromString(dbUser.ID).
		TelegramID(dbUser.TelegramID).
		FirstName(dbUser.FirstName).
		LastName(dbUser.LastName).
		Username(dbUser.Username).
		CreatedAt(createdAt).
		UpdatedAt(updatedAt).
		Build()

	if err != nil {
		return domainUser.User{}, fmt.Errorf("failed to build user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, telegramId domainUser.TelegramID, firstName domainUser.FirstName, lastName domainUser.LastName, username domainUser.Username) (domainUser.User, error) {
	q := New(r.db)

	now := time.Now()

	newUser, err := domainUser.NewBuilder().
		NewID().
		TelegramID(telegramId).
		FirstName(firstName).
		LastName(lastName).
		Username(username).
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	if err != nil {
		return domainUser.User{}, fmt.Errorf("failed to build user: %w", err)
	}

	_, err = q.CreateUser(ctx, CreateUserParams{
		ID:         newUser.ID().String(),
		TelegramID: newUser.TelegramID(),
		FirstName:  newUser.FirstName(),
		LastName:   newUser.LastName(),
		Username:   newUser.Username(),
		CreatedAt:  now.Format(SQLITE_TIME_FORMAT),
		UpdatedAt:  now.Format(SQLITE_TIME_FORMAT),
	})
	if err != nil {
		return domainUser.User{}, fmt.Errorf("failed to create user: %w", err)
	}
	return newUser, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user domainUser.User) (domainUser.User, error) {
	q := New(r.db)

	user = user.UpdateTimestamp()

	_, err := q.UpdateUser(ctx, UpdateUserParams{
		ID:         user.ID().String(),
		TelegramID: user.TelegramID(),
		FirstName:  user.FirstName(),
		LastName:   user.LastName(),
		Username:   user.Username(),
		UpdatedAt:  user.UpdatedAt().Format(SQLITE_TIME_FORMAT),
	})
	if err != nil {
		return domainUser.User{}, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}
