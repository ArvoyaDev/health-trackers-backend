package main

import (
	"encoding/json"
	"net/http"

	db "github.com/ArvoyaDev/health-trackers-backend/internal/mysql"
)

func (c *config) getUser(w http.ResponseWriter, r *http.Request) {
	// Retrieve claims from context
	claims, ok := r.Context().Value("User-claims").(map[string]interface{})
	if !ok {
		http.Error(w, "Claims not found", http.StatusUnauthorized)
		return
	}

	// Extract user info from claims
	sub, ok := claims["username"].(string)
	if !ok {
		http.Error(w, "username claim missing or invalid", http.StatusUnauthorized)
		return
	}

	// Connect to the database
	database, err := db.New()
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	// Check if user exists in the database
	user, err := database.GetUserBySub(sub)
	if err != nil {
		http.Error(w, "Failed to get user: "+err.Error(), http.StatusNotFound)
		return
	}

	// Define response struct with trackers and their symptoms
	type trackerResponse struct {
		ID          int             `json:"id"`
		TrackerName string          `json:"tracker_name"`
		Symptoms    []db.Symptom    `json:"symptoms"`
		Logs        []db.SymptomLog `json:"logs"`
	}

	type Response struct {
		Trackers []trackerResponse `json:"trackers"`
	}

	// Get trackers by user ID
	trackers, err := database.GetTrackerByUserID(user.ID)
	if err != nil {
		http.Error(w, "Failed to get trackers: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Initialize the response trackers slice
	var responseTrackers []trackerResponse

	// Loop through trackers and attach symptoms
	for _, tracker := range trackers {
		// Get symptoms for each tracker
		symptoms, err := database.GetSymptomsByTrackerID(tracker.ID)
		if err != nil {
			http.Error(w, "Failed to get symptoms: "+err.Error(), http.StatusInternalServerError)
			return
		}
		logs, err := database.GetSymptomLogsByTrackerID(tracker.ID)
		if err != nil {
			http.Error(w, "Failed to get logs: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Create tracker response with symptoms
		trackerRes := trackerResponse{
			ID:          tracker.ID,
			TrackerName: tracker.TrackerName,
			Symptoms:    symptoms, // Attach symptoms to the tracker
			Logs:        logs,
		}

		// Append to the response
		responseTrackers = append(responseTrackers, trackerRes)
	}

	// Build the final response
	response := Response{
		Trackers: responseTrackers,
	}

	// Convert the response to JSON
	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to serialize user data", http.StatusInternalServerError)
		return
	}

	// Send the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func (c *config) createUser(w http.ResponseWriter, r *http.Request) {
	// Retrieve claims from context
	claims, ok := r.Context().Value("User-claims").(map[string]interface{})
	if !ok {
		http.Error(w, "Claims not found", http.StatusUnauthorized)
		return
	}

	// Extract user info from claims
	sub, ok := claims["username"].(string)
	if !ok {
		http.Error(w, "username claim missing or invalid", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var user db.CompleteUser
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create user in the database
	database, err := db.New()
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	err = database.CreateUser(user.Email, sub)
	if err != nil {
		error := "Failed to create user: " + err.Error()
		http.Error(w, error, http.StatusInternalServerError)
		return
	}

	// Get the created user from the database
	createdUser, err := database.GetUserBySub(sub)
	if err != nil {
		error := "Failed to get user: " + err.Error()
		http.Error(w, error, http.StatusNotFound)
		return
	}

	err = database.CreateTracker(user.Tracker, createdUser.ID)
	if err != nil {
		error := "Failed to create tracker: " + err.Error()
		http.Error(w, error, http.StatusInternalServerError)
		return
	}

	createdTracker, err := database.GetTrackerByNameAndUserID(user.Tracker, createdUser.ID)
	if err != nil {
		error := "Failed to get tracker: " + err.Error()
		http.Error(w, error, http.StatusInternalServerError)
		return
	}

	for _, symptom := range user.Symptoms {
		err := database.CreateSymptom(symptom, createdTracker.ID)
		if err != nil {
			error := "Failed to create symptom: " + err.Error()
			http.Error(w, error, http.StatusInternalServerError)
			return
		}
	}

	type Response struct {
		UserID   int        `json:"user_id"`
		Tracker  db.Tracker `json:"tracker"`
		Symptoms []string   `json:"symptoms"`
	}

	response := Response{
		UserID:   createdUser.ID,
		Tracker:  createdTracker,
		Symptoms: user.Symptoms,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to serialize user data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonData)
}

func (c *config) createTracker(w http.ResponseWriter, r *http.Request) {
	// Retrieve claims from context
	claims, ok := r.Context().Value("User-claims").(map[string]interface{})
	if !ok {
		http.Error(w, "Claims not found", http.StatusUnauthorized)
		return
	}

	// Extract user info from claims
	sub, ok := claims["username"].(string)
	if !ok {
		http.Error(w, "username claim missing or invalid", http.StatusUnauthorized)
		return
	}

	database, err := db.New()
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	// Get User ID from the database
	user, err := database.GetUserBySub(sub)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check Tracker Count
	trackers, err := database.GetTrackerByUserID(user.ID)
	if err != nil {
		http.Error(w, "Failed to get trackers", http.StatusInternalServerError)
		return
	}

	if len(trackers) >= 5 {
		http.Error(w, "Tracker limit reached", http.StatusForbidden)
		return
	}

	// Parse request body
	var tracker db.NewTrackerRequestBody
	if err := json.NewDecoder(r.Body).Decode(&tracker); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	tracker.UserID = user.ID

	// Create tracker in the database
	err = database.CreateTracker(tracker.TrackerName, tracker.UserID)
	if err != nil {
		error := "Failed to create tracker: " + err.Error()
		http.Error(w, error, http.StatusConflict)
		return
	}

	// Get the created tracker from the database
	createdTracker, err := database.GetTrackerByNameAndUserID(tracker.TrackerName, tracker.UserID)
	if err != nil {
		error := "Failed to get tracker: " + err.Error()
		http.Error(w, error, http.StatusInternalServerError)
		return
	}

	for _, symptom := range tracker.Symptoms {
		err := database.CreateSymptom(symptom, createdTracker.ID)
		if err != nil {
			error := "Failed to create symptom: " + err.Error()
			http.Error(w, error, http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
}

func (c *config) createSymptoms(w http.ResponseWriter, r *http.Request) {
	// Retrieve claims from context
	claims, ok := r.Context().Value("User-claims").(map[string]interface{})
	if !ok {
		http.Error(w, "Claims not found", http.StatusUnauthorized)
		return
	}

	// Extract user info from claims
	sub, ok := claims["username"].(string)
	if !ok {
		http.Error(w, "username claim missing or invalid", http.StatusUnauthorized)
		return
	}

	database, err := db.New()
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	// Get User ID from the database
	user, err := database.GetUserBySub(sub)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Validate request
	if user.CognitoSub != sub {
		http.Error(w, "User ID mismatch", http.StatusBadRequest)
		return
	}

	type SymptomRequestBody struct {
		TrackerID int          `json:"tracker_id"`
		Symptoms  []db.Symptom `json:"symptoms"`
	}

	// Parse request body
	var symptoms SymptomRequestBody
	if err := json.NewDecoder(r.Body).Decode(&symptoms); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create symptoms in the database
	for _, symptom := range symptoms.Symptoms {
		symptom.TrackerID = symptoms.TrackerID
		err := database.CreateSymptom(symptom.SymptomName, symptom.TrackerID)
		if err != nil {
			http.Error(w, "Failed to create symptoms", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
}

func (c *config) createSymptomLog(w http.ResponseWriter, r *http.Request) {
	// Retrieve claims from context
	claims, ok := r.Context().Value("User-claims").(map[string]interface{})
	if !ok {
		http.Error(w, "Claims not found", http.StatusUnauthorized)
		return
	}

	// Extract user info from claims
	sub, ok := claims["username"].(string)
	if !ok {
		http.Error(w, "username claim missing or invalid", http.StatusUnauthorized)
		return
	}

	database, err := db.New()
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	// Get User ID from the database
	user, err := database.GetUserBySub(sub)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Parse request body
	var symptomLog db.SymptomLogRequestBody
	if err := json.NewDecoder(r.Body).Decode(&symptomLog); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	symptomLog.UserID = user.ID

	tracker, err := database.GetTrackerByNameAndUserID(symptomLog.TrackerName, user.ID)
	if err != nil {
		http.Error(w, "Tracker not found", http.StatusNotFound)
		return
	}

	symptomLog.TrackerID = tracker.ID

	// Create symptom log in the database
	err = database.CreateSymptomLog(symptomLog)
	if err != nil {
		error := "Failed to create symptom log: " + err.Error()
		http.Error(w, error, http.StatusInternalServerError)
		return
	}

	// Get the created symptom log from the database where time is the closest to the current time
	createdSymptomLog, err := database.GetSymptomLogByTrackerIDAndCurrentTime(tracker.ID)
	if err != nil {
		error := "Failed to get symptom log: " + err.Error()
		http.Error(w, error, http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(createdSymptomLog)
	if err != nil {
		http.Error(w, "Failed to serialize symptom log data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonData)
}

func (c *config) getSymptomLogs(w http.ResponseWriter, r *http.Request) {
	// Retrieve claims from context
	claims, ok := r.Context().Value("User-claims").(map[string]interface{})
	if !ok {
		http.Error(w, "Claims not found", http.StatusUnauthorized)
		return
	}

	// Extract user info from claims
	sub, ok := claims["username"].(string)
	if !ok {
		http.Error(w, "username claim missing or invalid", http.StatusUnauthorized)
		return
	}

	database, err := db.New()
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	// Get User ID from the database
	user, err := database.GetUserBySub(sub)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Validate request
	if user.CognitoSub != sub {
		http.Error(w, "User ID mismatch", http.StatusBadRequest)
		return
	}

	// Get symptom logs from the database
	symptomLogs, err := database.GetSymptomLogsByUserID(user.ID)
	if err != nil {
		http.Error(w, "Failed to get symptom logs", http.StatusInternalServerError)
		return
	}

	// Convert the response to JSON
	jsonData, err := json.Marshal(symptomLogs)
	if err != nil {
		http.Error(w, "Failed to serialize symptom logs", http.StatusInternalServerError)
		return
	}

	// Send the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
