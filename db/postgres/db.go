package postgres

import (
	"context"
	"database/sql"
	_"github.com/lib/pq"
	"time"
	"video-analytics-pipe/config"
)

func New() (*sql.DB, error) {

	dsn := config.GetString("DB_ADDR", "")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
