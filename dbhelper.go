package dbhelper

import (
	"context"
	"database/sql"
	_ "embed"
	"os"

	"bufo.zone/dbufo"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var tables string

func GetDb(ctx context.Context) *dbufo.Queries {
	dbPath := os.Getenv("DB_PATH")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}

	if _, err := db.ExecContext(ctx, tables); err != nil {
		panic(err)
	}

	return dbufo.New(db)
}
