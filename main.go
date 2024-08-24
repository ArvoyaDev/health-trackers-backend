package main

import (
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

type config struct {
	dbName         string
	dataSourceName string
}

func main() {
	// Load environment variables from .env file
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }
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
	corsMux := http.HandlerFunc(corsMiddleware(mux))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}
	log.Printf("Server listening on port %s", port)
	log.Fatal(srv.ListenAndServe())
}

func corsMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the Access-Control-Allow-Origin header to allow all origins
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Set the Access-Control-Allow-Methods header to allow all methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		// Set the Access-Control-Allow-Headers header to allow all headers
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// If the request method is OPTIONS, return a 200 OK status
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	}
}
