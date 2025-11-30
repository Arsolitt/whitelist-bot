package db

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func GetSqliteDB(ctx context.Context, url string) (*sql.DB, error) {
	return sql.Open("sqlite3", url)
}
