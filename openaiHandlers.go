package main

import (
	"encoding/json"
	"net/http"

	openai "github.com/ArvoyaDev/symptom-tracker-backend/internal/openai"
	_ "github.com/go-sql-driver/mysql"
)

func (cfg *config) openai(w http.ResponseWriter, r *http.Request) {
	var selectedTracker openai.SelectedTracker

	err := json.NewDecoder(r.Body).Decode(&selectedTracker)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	res, err := openai.Openaimain(
		selectedTracker.MedicalType,
		selectedTracker.Logs,
	)
	if err != nil {
		http.Error(
			w,
			"Failed to get response from OpenAI: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(res))
}
