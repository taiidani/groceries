package db

import (
	"context"
	"database/sql"
	"embed"

	_ "github.com/glebarez/go-sqlite"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var schema embed.FS

func NewMemory(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:") // ?_pragma=foreign_keys(1)
	if err != nil {
		return nil, err
	}

	return db, ensureSchema(ctx, db, "sqlite")
}

func NewFile(ctx context.Context, filename string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", filename+"") //?_pragma=foreign_keys(1)
	if err != nil {
		return nil, err
	}

	err = db.PingContext(ctx)
	if err != nil {
		return db, err
	}

	return db, ensureSchema(ctx, db, "sqlite")
}

func NewConn(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.PingContext(ctx)
	if err != nil {
		return db, err
	}

	return db, ensureSchema(ctx, db, "postgres")
}

func ensureSchema(_ context.Context, db *sql.DB, dialect string) error {
	goose.SetBaseFS(schema)

	if err := goose.SetDialect(dialect); err != nil {
		return err
	}

	return goose.Up(db, "migrations")
}
