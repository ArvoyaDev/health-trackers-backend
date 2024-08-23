package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// NOTE: Database holds the database connection pool.
type Database struct {
	mysql *sql.DB
}

// NOTE: New initializes a new Database connection.
func New(dsn string) (*Database, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// NOTE: Set database connection parameters
	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	// NOTE: Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	log.Println("Connected to the database successfully")
	return &Database{db}, nil
}

// NOTE: Close closes the database connection.
func (d *Database) Close() error {
	fmt.Println("Closing database connection")
	return d.mysql.Close()
}

func (d *Database) SelectDatabase(database string) error {
	_, err := d.mysql.Exec("USE " + database)
	if err != nil {
		return fmt.Errorf("error selecting database: %w", err)
	}
	fmt.Println("Selected database: " + database)
	return nil
}
