package main

import (
	"encoding/json"
	"log"
	"net/http"

	db "github.com/ArvoyaDev/symptom-tracker-backend/internal/mysql"
)

func (cfg *config) getHeartburnLogs(w http.ResponseWriter, r *http.Request) {
	database, err := db.New(cfg.dataSourceName)
	if err != nil {
		log.Printf("error creating database connection: %v", err)
		http.Error(w, "error creating database connection", http.StatusInternalServerError)
		return
	}
	err = database.SelectDatabase(cfg.dbName)
	if err != nil {
		log.Printf("error selecting database: %v", err)
		http.Error(w, "error selecting database", http.StatusInternalServerError)
		return
	}
	defer database.Close()
	logs, err := database.GetHeartburnLogs()
	if err != nil {
		log.Printf("error getting logs: %v", err)
		http.Error(w, "error getting logs", http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(logs)
	if err != nil {
		log.Printf("error marshalling logs: %v", err)
		http.Error(w, "error marshalling logs", http.StatusInternalServerError)
		return
	}
	// Set the Content-Type to text/html
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonData))
}

func (cfg *config) createHeartburnLog(w http.ResponseWriter, r *http.Request) {
	database, err := db.New(cfg.dataSourceName)
	if err != nil {
		log.Printf("error creating database connection: %v", err)
		http.Error(w, "error creating database connection", http.StatusInternalServerError)
		return
	}
	err = database.SelectDatabase(cfg.dbName)
	if err != nil {
		log.Printf("error selecting database: %v", err)
		http.Error(w, "error selecting database", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	var symptom db.HeartburnLog

	if err := json.NewDecoder(r.Body).Decode(&symptom); err != nil {
		log.Printf("error decoding request body: %v", err)
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}
	if err := database.CreateHeartburnLog(symptom); err != nil {
		log.Printf("error creating log: %v", err)
		http.Error(w, "error creating log", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
