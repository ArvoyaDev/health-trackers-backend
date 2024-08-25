package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ArvoyaDev/symptom-tracker-backend/internal/auth"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

type config struct {
	dbName         string
	dataSourceName string
	AuthClient     *auth.CognitoClient
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	dataSourceName := os.Getenv("AWS_DATABASE_URL")
	port := os.Getenv("PORT")
	dbName := os.Getenv("DATABASE_NAME")
	authClient := auth.Init()
	config := config{
		dbName:         dbName,
		dataSourceName: dataSourceName,
		AuthClient:     authClient,
	}

	// Main router with subrouting
	mainMux := http.NewServeMux()

	// DB Mux & routes
	dbMux := http.NewServeMux()

	mainMux.Handle("/db/", http.StripPrefix("/db", dbMux))

	dbMux.HandleFunc("GET /logs", config.getHeartburnLogs)
	dbMux.HandleFunc("POST /logs", config.createHeartburnLog)

	// Cognito Mux & routes
	cognitoMux := http.NewServeMux()

	mainMux.Handle("/aws-cognito/", http.StripPrefix("/aws-cognito", cognitoMux))

	cognitoMux.HandleFunc("POST /signup", config.signUp)
	cognitoMux.HandleFunc("POST /confirm-signup", config.ConfirmSignup)
	cognitoMux.HandleFunc("POST /request-verification-code", config.RequestVerificationCode)

	// Apply CORS middleware
	corsMux := corsMiddleware(mainMux)

	// Apply rate limiter middleware
	rateLimitMux := rateLimitMiddleware(corsMux)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: rateLimitMux,
	}
	log.Printf("Server listening on port %s", port)
	log.Fatal(srv.ListenAndServe())
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set the Access-Control-Allow-Origin header to allow all origins
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Set the Access-Control-Allow-Methods header to allow all methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		// Set the Access-Control-Allow-Headers header to allow all headers
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func rateLimitMiddleware(next http.Handler) http.HandlerFunc {
	// Set the rate limit to 15 requests per second with a burst of 5 request
	limiter := rate.NewLimiter(rate.Limit(15), 5)

	return func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	}
}
