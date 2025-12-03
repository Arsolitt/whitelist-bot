package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"whitelist-bot/internal/core"

	domainUser "whitelist-bot/internal/domain/user"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type iQueryable interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, optionsAndArgs ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, optionsAndArgs ...any) pgx.Row
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

	user, err := domainUser.NewBuilder().
		ID(dbUser.ID).
		TelegramID(dbUser.TelegramID).
		ChatID(dbUser.ChatID).
		FirstName(dbUser.FirstName).
		LastName(dbUser.LastName).
		Username(dbUser.Username).
		CreatedAt(dbUser.CreatedAt).
		UpdatedAt(dbUser.UpdatedAt).
		Build()

	if err != nil {
		return domainUser.User{}, fmt.Errorf("failed to build user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) CreateUser(
	ctx context.Context,
	telegramId domainUser.TelegramID,
	chatID domainUser.ChatID,
	firstName domainUser.FirstName,
	lastName domainUser.LastName,
	username domainUser.Username,
) (domainUser.User, error) {
	q := New(r.db)

	now := time.Now()

	newUser, err := domainUser.NewBuilder().
		NewID().
		TelegramID(telegramId).
		ChatID(chatID).
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
		ID:         newUser.ID(),
		TelegramID: newUser.TelegramID(),
		ChatID:     newUser.ChatID(),
		FirstName:  newUser.FirstName(),
		LastName:   newUser.LastName(),
		Username:   newUser.Username(),
		CreatedAt:  now,
		UpdatedAt:  now,
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
		ID:         user.ID(),
		TelegramID: user.TelegramID(),
		ChatID:     user.ChatID(),
		FirstName:  user.FirstName(),
		LastName:   user.LastName(),
		Username:   user.Username(),
		UpdatedAt:  user.UpdatedAt(),
	})
	if err != nil {
		return domainUser.User{}, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) UserByID(ctx context.Context, id domainUser.ID) (domainUser.User, error) {
	q := New(r.db)

	dbUser, err := q.UserByID(ctx, id)
	if err != nil {
		return domainUser.User{}, fmt.Errorf("failed to get user by id: %w", err)
	}

	user, err := domainUser.NewBuilder().
		ID(dbUser.ID).
		TelegramID(dbUser.TelegramID).
		ChatID(dbUser.ChatID).
		FirstName(dbUser.FirstName).
		LastName(dbUser.LastName).
		Username(dbUser.Username).
		CreatedAt(dbUser.CreatedAt).
		UpdatedAt(dbUser.UpdatedAt).
		Build()
	if err != nil {
		return domainUser.User{}, fmt.Errorf("failed to build user: %w", err)
	}
	return user, nil
}
