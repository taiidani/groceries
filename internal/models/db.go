package models

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"

	internalDB "github.com/taiidani/groceries/internal/db"
)

var db *sql.DB

func InitDB(ctx context.Context) error {
	switch os.Getenv("DB_TYPE") {
	case "postgres":
		client, err := internalDB.New(ctx, os.Getenv("DATABASE_URL"))
		db = client
		return err
	default:
		return errors.New("unknown DB_TYPE database version specified")
	}
}

func insertWithID(ctx context.Context, tx *sql.Tx, query string, args ...any) (int, error) {
	var id int

	if !strings.Contains(query, "RETURNING") {
		return id, errors.New("inserting with ID requires the use of the RETURNING directive")
	}

	err := tx.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil {
		return id, errors.Join(tx.Rollback(), err)
	}

	return id, nil
}
