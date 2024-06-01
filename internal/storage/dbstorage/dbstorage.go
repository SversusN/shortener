package dbstorage

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresDB struct {
	db *sql.DB
}

func NewDb(connectionString string) (*PostgresDB, error) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection to postgresql: %w", err)
	}
	//defer func(db *sql.DB) {
	//	err := db.Close()
	//	if err != nil {
	//		log.Fatalf("failed to close connection to postgresql: %w", err)
	//	}
	//}(db)
	err = db.Ping()
	if err != nil {
		err := db.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to close PostgreSQL connection after db.Ping: %w", err)
		}
		return nil, fmt.Errorf("failed to ping PostgreSQL connection: %w", err)
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

func (pg *PostgresDB) GetURL(id string) (string, error) {
	return id, nil
}

func (pg *PostgresDB) SetURL(id string, targetURL string) error {
	return nil
}

func (pg *PostgresDB) Ping() error {
	err := pg.db.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}
	return nil
}
