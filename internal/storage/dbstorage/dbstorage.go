package dbstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresDB struct {
	db *sql.DB
}

func NewDB(connectionString string) (*PostgresDB, error) {
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
	_, err = db.ExecContext(context.Background(), "CREATE TABLE IF NOT EXISTS URLS (short_url varchar(100), original_url varchar(1000));"+
		"CREATE UNIQUE INDEX IF NOT EXISTS original_url_idx ON URLS (original_url);")
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL table: %w", err)
	}
	return &PostgresDB{
		db: db,
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
	row := pg.db.QueryRowContext(context.Background(), query, shortURL)
	var originalURL string
	err := row.Scan(&originalURL)
	if err != nil {
		return "", fmt.Errorf("failed to query short URL: %w", err)
	}
	return originalURL, nil
}

func (pg *PostgresDB) SetURL(id string, originalUrl string) error {
	query := "INSERT INTO URLS (short_url, original_url) VALUES ($1, $2)"
	_, err := pg.db.ExecContext(context.Background(), query, id, originalUrl)
	if err != nil {
		return fmt.Errorf("failed to insert URL: %w", err)
	}
	return nil
}

func (pg *PostgresDB) SetURLBatch(ctx context.Context, u map[string]string) error {
	tx, err := pg.db.Begin()
	if err != nil {
		return err
	}
	query := "INSERT INTO URLS (short_url, original_url) VALUES ($1, $2)"
	stmt, _ := tx.PrepareContext(ctx, query)
	defer stmt.Close()
	for s := range u {
		_, err := stmt.ExecContext(context.Background(), s, u[s])
		if err != nil {
			tx.Rollback()
			return errors.New("Doubled key in batch. Not allowed")
		}
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (pg *PostgresDB) GetKey(originalURL string) (string, error) {
	var storedURL string
	rowExist := pg.db.QueryRowContext(
		context.Background(),
		`SELECT short_url FROM URLS WHERE original_url=$1 LIMIT 1`,
		originalURL)
	err := rowExist.Scan(&storedURL)
	if err != nil {
		return "", err
	}
	if originalURL == "" {
		return "", errors.New("nothing found for short URL")
	}
	return storedURL, nil
}

func (pg *PostgresDB) Ping() error {
	err := pg.db.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}
	return nil
}
