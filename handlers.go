package main

import (
	"encoding/json"
	"net/http"

	db "github.com/ArvoyaDev/symptom-tracker-backend/internal/mysql"
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
		http.Error(w, "usernmae claim missing or invalid", http.StatusUnauthorized)
		return
	}

	database, err := db.New()
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}

	defer database.Close()
	// Check if user exists in the database
	user, err := database.GetUserBySub(sub)
	if err != nil {
		error := "Failed to get user: " + err.Error()
		http.Error(w, error, http.StatusNotFound)
		return
	}
	illnesses, err := database.GetIllnessesByUserID(user.ID)
	if err != nil {
		error := "Failed to get illnesses: " + err.Error()
		http.Error(w, error, http.StatusInternalServerError)
		return
	}

	allSymptoms := make([][]db.Symptom, len(illnesses))

	for i, illness := range illnesses {
		symptoms, err := database.GetSymptomsByIllnessID(illness.ID)
		if err != nil {
			error := "Failed to get symptoms: " + err.Error()
			http.Error(w, error, http.StatusInternalServerError)
			return
		}
		allSymptoms[i] = symptoms
	}

	// Respond with user profile
	response := map[string]interface{}{
		"username":  user.CognitoSub,
		"email":     user.Email,
		"illnesses": illnesses,
		"symptoms":  allSymptoms,
		// Include other user details as needed
	}
	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to serialize user data", http.StatusInternalServerError)
		return
	}

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
	var user db.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if user.CognitoSub != sub {
		http.Error(w, "User ID mismatch", http.StatusBadRequest)
		return
	}

	// Create user in the database
	database, err := db.New()
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	err = database.CreateUser(user)
	if err != nil {
		error := "Failed to create user: " + err.Error()
		http.Error(w, error, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (c *config) createIllness(w http.ResponseWriter, r *http.Request) {
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

	// Parse request body
	var illness db.Illness
	if err := json.NewDecoder(r.Body).Decode(&illness); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	illness.UserID = user.ID

	// Create illness in the database
	err = database.CreateIllness(illness)
	if err != nil {
		http.Error(w, "Failed to create illness", http.StatusInternalServerError)
		return
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
		IllnessID int          `json:"illness_id"`
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
		symptom.IllnessID = symptoms.IllnessID
		err := database.CreateSymptom(symptom)
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

	// Validate request
	if user.CognitoSub != sub {
		http.Error(w, "User ID mismatch", http.StatusBadRequest)
		return
	}

	// Parse request body
	var symptomLog db.SymptomLog
	if err := json.NewDecoder(r.Body).Decode(&symptomLog); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create symptom log in the database
	err = database.CreateSymptomLog(symptomLog)
	if err != nil {
		http.Error(w, "Failed to create symptom log", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
