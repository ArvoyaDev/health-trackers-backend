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
	sub, ok := claims["cognito:username"].(string)
	if !ok {
		http.Error(w, "cognito:usernmae claim missing or invalid", http.StatusUnauthorized)
		return
	}

	// Check if user exists in the database
	user, err := c.db.GetUserBySub(sub)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Respond with user profile
	response := map[string]interface{}{
		"username": user.CognitoSub,
		"email":    user.Email,
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

	err := c.db.CreateUser(user)
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

	// Get User ID from the database
	user, err := c.db.GetUserBySub(sub)
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
	err = c.db.CreateIllness(illness)
	if err != nil {
		http.Error(w, "Failed to create illness", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
