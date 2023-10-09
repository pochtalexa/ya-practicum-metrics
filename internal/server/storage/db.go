package storage

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pochtalexa/ya-practicum-metrics/internal/server/flags"
	"time"
)

func InitDB() (*sql.DB, error) {
	ps := flags.FlagDBConn

	db, err := sql.Open("pgx", ps)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func PingDB(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return err
	}

	return nil
}
