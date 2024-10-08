// Пакет dbstorage описывает поведение БД
package dbstorage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/SversusN/shortener/internal/internalerrors"
	utils "github.com/SversusN/shortener/internal/pkg/migrator"
)

// PostgresDB структура для реализации sql.DB с помощью драйвера для PostgreSQL
type PostgresDB struct {
	ctx context.Context
	db  *sql.DB
}

//go:embed migrations/*.sql
var MigrationsFS embed.FS

// migrationsDir - локальная папка с миграциями
const migrationsDir = "migrations"

// NewDB конструктор для объекта БД
func NewDB(ctx context.Context, connectionString string) (*PostgresDB, error) {
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

	migrator := utils.MustGetNewMigrator(MigrationsFS, migrationsDir)
	err = migrator.ApplyMigrations(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create table URLs: %w", err)
	}
	log.Println("Migrations applied!")
	return &PostgresDB{
		db:  db,
		ctx: ctx,
	}, nil
}

// Close -метод закрытия соединения
func (pg *PostgresDB) Close() {
	if pg.db != nil {
		err := pg.db.Close()
		if err != nil {
			log.Fatalf("Error closing database connection: %v\n", err)
		}
		log.Println("Database connection closed.")
	}
}

// GetURL - реализация метода получения единичной ссылки
func (pg *PostgresDB) GetURL(shortURL string) (string, error) {
	query := "SELECT original_url, COALESCE(is_deleted, FALSE) as is_deleted FROM URLS WHERE short_url=$1"
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

// SetURL реализация метода сохранения едичничной ссылки
func (pg *PostgresDB) SetURL(shortURL string, originalURL string, userID string) (string, error) {
	if userID == "" {
		userID = uuid.Nil.String()
	}

	tx, err := pg.db.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	var keyExist string
	queryCheck := "SELECT short_url FROM URLS WHERE original_url=$1 LIMIT 1 FOR UPDATE"
	query := "INSERT INTO URLS (short_url, original_url, user_id) VALUES ($1, $2, $3)"
	errKeyExist := tx.QueryRowContext(pg.ctx, queryCheck, originalURL).Scan(&keyExist)
	if errors.Is(errKeyExist, sql.ErrNoRows) {
		tx.QueryRowContext(pg.ctx, query, shortURL, originalURL, userID)
		tx.Commit()
		return shortURL, nil
	} else {
		tx.Rollback()
		return keyExist, internalerrors.ErrOriginalURLAlreadyExists
	}
}

// SetURLBatch сохранение массива ссылок
func (pg *PostgresDB) SetURLBatch(u map[string]UserURL) (map[string]UserURL, error) {
	result := make(map[string]UserURL)
	tx, err := pg.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	queryCheck := "SELECT short_url FROM URLS WHERE original_url=$1 LIMIT 1 FOR UPDATE"
	query := "INSERT INTO URLS (short_url, original_url, user_id) VALUES ($1, $2, $3)"
	var possibleError error
	for s := range u {
		var keyExist string
		errBlankKey := tx.QueryRowContext(pg.ctx, queryCheck, u[s].OriginalURL).Scan(&keyExist)
		if errors.Is(errBlankKey, sql.ErrNoRows) {
			tx.QueryRowContext(pg.ctx, query, s, u[s].OriginalURL, u[s].UserID)
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

// Ping - метод проверки соединения с БД Postgre
func (pg *PostgresDB) Ping() error {
	err := pg.db.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}
	return nil
}

// GetUserUrls получение массива ссылок  с фильтром пользователя
func (pg *PostgresDB) GetUserUrls(userID string) (any, error) {
	result := make([]UserURLEntity, 0)
	query := "SELECT short_url, original_url FROM URLS WHERE user_id = $1 and is_deleted = FALSE;"
	rows, err := pg.db.QueryContext(pg.ctx, query, userID)
	if err != nil {
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, internalerrors.ErrNotFound
	}
	return result, err
}

// DeleteUserURLs реализация асинхронного удаления ссылок по ИД пользователя
func (pg *PostgresDB) DeleteUserURLs(userID string, group *sync.WaitGroup) (deletedURLs chan string, err error) {
	deletedURLs = make(chan string)
	tx, err := pg.db.BeginTx(pg.ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("tran error: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	query := "UPDATE URLS set is_deleted = true WHERE user_id = $1 AND short_url = ANY($2);"
	group.Add(1)
	go func() {
		var forDelete []string //= make([]string, 0, 10)
		defer group.Done()
	chanloop:
		for {
			select {
			case key, ok := <-deletedURLs:
				{
					if !ok {
						break chanloop
					}
					forDelete = append(forDelete, key)
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
		_, err = tx.ExecContext(pg.ctx, query, userID, forDelete)
		if err != nil {
			return
		}
		err := tx.Commit()
		if err != nil {
			return
		}
	}()

	return deletedURLs, nil
}

// GetStats получение количества ссылок и уникальных пользователей
func (pg *PostgresDB) GetStats() (usersCount int, URLsCount int, statError error) {

	tx, err := pg.db.Begin()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	query := "SELECT COALESCE(count(*),0) as URLsCount FROM URLS WHERE is_deleted = FALSE;"
	row := tx.QueryRowContext(pg.ctx, query)
	err = row.Scan(URLsCount)
	if err != nil {
		return 0, 0, err
	}
	queryUsers := "SELECT COALESCE(count(distinct user_id),0) as usersCount FROM URLS WHERE is_deleted = FALSE;"
	rowUsers := tx.QueryRowContext(pg.ctx, queryUsers)
	err = rowUsers.Scan(usersCount)
	if err != nil {
		return 0, 0, err
	}
	tx.Commit()
	return usersCount, URLsCount, nil
}
