package dbstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/SversusN/shortener/internal/internalerrors"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresDB struct {
	db  *sql.DB
	ctx context.Context
}

func NewDB(connectionString string, ctx context.Context) (*PostgresDB, error) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection to postgresql: %w", err)
	}
	err = db.Ping()
	if err != nil {
		err := db.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to close PostgreSQL connection after db.Ping: %w", err)
		}
		return nil, fmt.Errorf("failed to ping PostgreSQL connection: %w", err)
	}
	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS URLS (short_url varchar(100), original_url varchar(1000));"+
		"CREATE UNIQUE INDEX IF NOT EXISTS original_url_idx ON URLS (original_url);")
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL table: %w", err)
	}
	return &PostgresDB{
		db:  db,
		ctx: ctx,
	}, nil
}

func (pg *PostgresDB) Close() {
	if pg.db != nil {
		err := pg.db.Close()
		if err != nil {
			log.Fatalf("Error closing database connection: %v\n", err)
		}
		log.Println("Database connection closed.")
	}
}

func (pg *PostgresDB) GetURL(shortURL string) (string, error) {
	query := "SELECT original_url FROM URLS WHERE short_url=$1"
	row := pg.db.QueryRowContext(pg.ctx, query, shortURL)
	var originalURL string
	err := row.Scan(&originalURL)
	if err != nil {
		return "", fmt.Errorf("failed to query short URL: %w", err)
	}
	return originalURL, nil
}

func (pg *PostgresDB) SetURL(shortURL string, originalURL string) (string, error) {
	tx, err := pg.db.Begin()
	if err != nil {
		tx.Rollback()
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	var keyExist string
	queryCheck := "SELECT short_url FROM URLS WHERE original_url=$1 LIMIT 1 FOR UPDATE"
	query := "INSERT INTO URLS (short_url, original_url) VALUES ($1, $2)"
	errKeyExist := tx.QueryRowContext(pg.ctx, queryCheck, originalURL).Scan(&keyExist)
	if errors.Is(errKeyExist, sql.ErrNoRows) {
		err := tx.QueryRowContext(pg.ctx, query, shortURL, originalURL)
		if err != nil {
			fmt.Errorf("failed to query short URL: %w", err)
		}
		tx.Commit()
		return shortURL, nil
	} else {
		tx.Commit()
		return keyExist, internalerrors.ErrOriginalURLAlreadyExists
	}
}

func (pg *PostgresDB) SetURLBatch(u map[string]string) (map[string]string, error) {
	result := make(map[string]string)
	tx, err := pg.db.Begin()
	queryCheck := "SELECT short_url FROM URLS WHERE original_url=$1 LIMIT 1 FOR UPDATE"
	query := "INSERT INTO URLS (short_url, original_url) VALUES ($1, $2)"
	var possibleError error
	for s := range u {
		var keyExist string
		errBlankKey := tx.QueryRowContext(pg.ctx, queryCheck, u[s]).Scan(&keyExist)
		if errors.Is(errBlankKey, sql.ErrNoRows) {
			err := tx.QueryRowContext(pg.ctx, query, s, u[s])
			if err != nil {
				fmt.Errorf("failed to insert URL: %w", err)
			}
			result[s] = u[s]
		} else {
			possibleError = internalerrors.ErrOriginalURLAlreadyExists
			result[keyExist] = u[s]
		}
	}
	err = tx.Commit()
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return result, possibleError
}

func (pg *PostgresDB) Ping() error {
	err := pg.db.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}
	return nil
}
