package dbhelper

import (
	"context"
	_ "embed"
	"os"

	"bufo.zone/dbufo"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DbHelpers struct {
	Pool    *pgxpool.Pool
	Queries *dbufo.Queries
}

func GetDb(ctx context.Context) *DbHelpers {
	dbUrl := os.Getenv("DB_URL")
	db, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		panic(err)
	}

	return &DbHelpers{
		Pool:    db,
		Queries: dbufo.New(db),
	}
}

func (helpers *DbHelpers) Close() {
	helpers.Pool.Close()
}

func GetConn(ctx context.Context) *pgx.Conn {
	// dbPath := os.Getenv("DB_PATH")
	dbUrl := "postgres://postgres:password@localhost:5432/bufozone"
	db, err := pgx.Connect(ctx, dbUrl)
	if err != nil {
		panic(err)
	}

	return db
}
