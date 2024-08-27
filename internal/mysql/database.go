package db

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
)

// NOTE: Database holds the database connection pool.
type Database struct {
	mysql *sql.DB
}

type DBClientData struct {
	AwsRegion   string
	DbName      string
	DbUser      string
	RdsEndpoint string
}

// NOTE: New initializes a new Database connection.
func New() (*Database, error) {
	// Retrieve database connection details from environment variables

	dbName := os.Getenv("DATABASE_NAME")
	dbUser := os.Getenv("DATABASE_USER")
	dbHost := os.Getenv("RDS_ENDPOINT")
	dbPort := 3306
	region := os.Getenv("AWS_REGION")

	// Construct the DB endpoint
	dbEndpoint := fmt.Sprintf("%s:%d", dbHost, dbPort)

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("configuration error: %w", err)
	}

	// Generate IAM authentication token
	authToken, err := auth.BuildAuthToken(
		context.TODO(), dbEndpoint, region, dbUser, cfg.Credentials,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create authentication token: %w", err)
	}

	// Load CA certificate
	caCert, err := os.ReadFile("./us-west-2-bundle.pem")
	if err != nil {
		return nil, fmt.Errorf("error reading CA certificate: %w", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Configure TLS settings
	tlsConfig := &tls.Config{
		RootCAs:    caCertPool,
		MinVersion: tls.VersionTLS12,
	}

	err = mysql.RegisterTLSConfig("custom", tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("error registering custom TLS config: %w", err)
	}

	// Create DSN with IAM authentication token and cleartext password support
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?tls=custom&allowCleartextPasswords=true",
		dbUser,
		authToken,
		dbEndpoint,
		dbName,
	)

	// Open the database connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Set database connection parameters
	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	log.Println("Connected to the database successfully")
	return &Database{db}, nil
}

// NOTE: Close closes the database connection.
func (d *Database) Close() error {
	log.Println("Closing database connection")
	return d.mysql.Close()
}
