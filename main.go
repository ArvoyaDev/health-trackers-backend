package main

import (
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type config struct {
	dbName         string
	dataSourceName string
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	data_source_name := os.Getenv("AWS_DATABASE_URL")
	port := os.Getenv("PORT")
	db_name := os.Getenv("DATABASE_NAME")
	config := config{
		dbName:         db_name,
		dataSourceName: data_source_name,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /logs", config.getHeartburnLogs)
	mux.HandleFunc("POST /logs", config.createHeartburnLog)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	log.Printf("Server listening on port %s", port)
	log.Fatal(srv.ListenAndServe())
}
