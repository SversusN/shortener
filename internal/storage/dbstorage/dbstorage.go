package dbstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/SversusN/shortener/internal/internalerrors"
)

type PostgresDB struct {
	db  *sql.DB
	ctx context.Context
}

func NewDB(connectionString string, ctx *context.Context) (*PostgresDB, error) {
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
	_, err = db.ExecContext(*ctx, `
		CREATE TABLE IF NOT EXISTS URLS 
		(short_url varchar(100) NOT NULL,
		original_url varchar(1000) NOT NULL,
		user_id INT,
		is_deleted BOOL default FALSE);
		CREATE UNIQUE INDEX IF NOT EXISTS original_url_idx ON URLS (original_url);`)
	if err != nil {
		return nil, fmt.Errorf("failed to create table URLs: %w", err)
	}
	_, err = db.ExecContext(*ctx, `
        CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY
        );`)
	if err != nil {
		return nil, fmt.Errorf("postgres exec (create users): %w", err)
	}

	return &PostgresDB{
		db:  db,
		ctx: *ctx,
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
	query := "SELECT original_url, is_deleted FROM URLS WHERE short_url=$1"
	row := pg.db.QueryRowContext(pg.ctx, query, shortURL)
	var (
		originalURL string
		isDeleted   bool
	)
	err := row.Scan(&originalURL, &isDeleted)
	if err != nil {
		return "", fmt.Errorf("failed to query short URL: %w", err)
	}
	if isDeleted {
		return "", internalerrors.ErrDeleted
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
		tx.QueryRowContext(pg.ctx, query, shortURL, originalURL)
		tx.Commit()
		return shortURL, nil
	} else {
		tx.Commit()
		return keyExist, internalerrors.ErrOriginalURLAlreadyExists
	}
}

func (pg *PostgresDB) SetUserURL(shortURL string, originalURL string, userID int) (string, error) {
	tx, err := pg.db.Begin()
	if err != nil {
		tx.Rollback()
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	var keyExist string
	queryCheck := "SELECT short_url FROM URLS WHERE original_url=$1 LIMIT 1 FOR UPDATE"
	query := "INSERT INTO URLS (short_url, original_url, user_id) VALUES ($1, $2, $3)"
	errKeyExist := tx.QueryRowContext(pg.ctx, queryCheck, originalURL).Scan(&keyExist)
	if errors.Is(errKeyExist, sql.ErrNoRows) {
		tx.QueryRowContext(pg.ctx, query, shortURL, originalURL, userID)
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
	if err != nil {
		tx.Rollback()
	}
	queryCheck := "SELECT short_url FROM URLS WHERE original_url=$1 LIMIT 1 FOR UPDATE"
	query := "INSERT INTO URLS (short_url, original_url) VALUES ($1, $2)"
	var possibleError error
	for s := range u {
		var keyExist string
		errBlankKey := tx.QueryRowContext(pg.ctx, queryCheck, u[s]).Scan(&keyExist)
		if errors.Is(errBlankKey, sql.ErrNoRows) {
			tx.QueryRowContext(pg.ctx, query, s, u[s])
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

func (pg *PostgresDB) SetUserURLBatch(u map[string]string, userID int) (map[string]string, error) {
	result := make(map[string]string)
	tx, err := pg.db.Begin()
	if err != nil {
		tx.Rollback()
	}
	queryCheck := "SELECT short_url FROM URLS WHERE original_url=$1 LIMIT 1 FOR UPDATE"
	query := "INSERT INTO URLS (short_url, original_url, user_id) VALUES ($1, $2, $3)"
	var possibleError error
	for s := range u {
		var keyExist string
		errBlankKey := tx.QueryRowContext(pg.ctx, queryCheck, u[s]).Scan(&keyExist)
		if errors.Is(errBlankKey, sql.ErrNoRows) {
			tx.QueryRowContext(pg.ctx, query, s, u[s], userID)
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

func (pg *PostgresDB) CreateUser(ctx context.Context) (int, error) {
	//https://stackoverflow.com/questions/55923755/postgres-insert-without-any-values-for-columns-all-are-default
	query := "INSERT INTO USERS DEFAULT VALUES RETURNING id;"
	var userID int
	err := pg.db.QueryRowContext(ctx, query).Scan(&userID)
	if err != nil {
		return 0, errors.New("postgres create userId")
	}
	return userID, nil
}

func (pg *PostgresDB) GetUserUrls(userID int) (any, error) {
	result := make([]UserURLEntity, 0)
	query := "SELECT short_url, original_url FROM URLS WHERE user_id = $1 and is_deleted = FALSE;"
	rows, err := pg.db.QueryContext(pg.ctx, query, userID)
	if err != nil || rows.Err() != nil {
		return nil, errors.New("error postgres get userUrls")
	}
	defer rows.Close()
	var count = 0
	for rows.Next() {
		count++
		resultRow := UserURLEntity{}
		err = rows.Scan(&resultRow.ShortURL, &resultRow.OriginalURL)
		if err != nil {
			log.Printf("postgres get userUrls: %v", err)
			return nil, err
		}
		result = append(result, resultRow)
	}
	if count == 0 {
		return nil, internalerrors.ErrNotFound
	}
	return result, err
}

func (pg *PostgresDB) DeleteUserURLs(userID int) (deletedURLs chan string, err error) {

	deletedURLs = make(chan string)
	tx, err := pg.db.BeginTx(pg.ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("tran error: %w", err)
	}
	query := "UPDATE URLS SET is_deleted = true WHERE short_url = $1 AND user_id = $2;"
	go func() {
	chanloop:
		for {
			select {
			case key, ok := <-deletedURLs:
				{
					if !ok {
						break chanloop

					}
					_, err = tx.ExecContext(pg.ctx, query, key, userID)
					if err != nil {
						tx.Rollback()
						return
					}
				}
			case <-pg.ctx.Done(): //if cancel in future
				{
					tx.Rollback()
					break chanloop
				}
			}
		}
		tx.Commit()
	}()

	return deletedURLs, nil
}
