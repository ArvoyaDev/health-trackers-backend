package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/ArvoyaDev/symptom-tracker-backend/internal/auth"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (cfg *config) signUp(w http.ResponseWriter, r *http.Request) {
	// Ensure it's a POST request
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var user User
	// Decode the JSON request body into the User struct
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Call the SignUp method from CognitoClient
	err := cfg.AuthClient.SignUp(context.Background(), user.Username, user.Password)
	if err != nil {
		http.Error(w, "Failed to sign up user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User signed up successfully"))
}

type ConfirmSignupRequest struct {
	Username         string `json:"username"`
	ConfirmationCode string `json:"confirmationCode"`
}

func (c *config) ConfirmSignup(w http.ResponseWriter, r *http.Request) {
	var req ConfirmSignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	secretHash, err := auth.CalculateSecretHash(
		c.AuthClient.AppClientID,
		os.Getenv("COGNITO_CLIENT_SECRET"),
		req.Username,
	)
	if err != nil {
		http.Error(w, "Failed to calculate secret hash", http.StatusInternalServerError)
		log.Printf("Failed to calculate secret hash: %v", err)
		return
	}

	_, err = c.AuthClient.ConfirmSignUp(
		context.TODO(),
		&cognitoidentityprovider.ConfirmSignUpInput{
			ClientId:         &c.AuthClient.AppClientID,
			Username:         &req.Username,
			SecretHash:       &secretHash,
			ConfirmationCode: &req.ConfirmationCode,
		},
	)
	if err != nil {
		http.Error(w, "Failed to confirm signup", http.StatusInternalServerError)
		log.Printf("Failed to confirm signup: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *config) RequestVerificationCode(w http.ResponseWriter, r *http.Request) {
	var req ConfirmSignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	secretHash, err := auth.CalculateSecretHash(
		c.AuthClient.AppClientID,
		os.Getenv("COGNITO_CLIENT_SECRET"),
		req.Username,
	)
	if err != nil {
		http.Error(w, "Failed to calculate secret hash", http.StatusInternalServerError)
		log.Printf("Failed to calculate secret hash: %v", err)
		return
	}

	_, err = c.AuthClient.ResendConfirmationCode(
		context.TODO(),
		&cognitoidentityprovider.ResendConfirmationCodeInput{
			SecretHash: &secretHash,
			ClientId:   &c.AuthClient.AppClientID,
			Username:   &req.Username,
		},
	)
	if err != nil {
		http.Error(w, "Failed to resend confirmation code", http.StatusInternalServerError)
		log.Printf("Failed to resend confirmation code: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Verification code resent successfully"))
}

type SignInResponse struct {
	AccessToken  *string `json:"accessToken"`
	ExpiresIn    int32   `json:"expiresIn"`
	TokenType    *string `json:"tokenType"`
	IDToken      *string `json:"idToken"`
	RefreshToken *string `json:"refreshToken"`
}

func (c *config) SignIn(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	secretHash, err := auth.CalculateSecretHash(
		c.AuthClient.AppClientID,
		os.Getenv("COGNITO_CLIENT_SECRET"),
		user.Username,
	)
	if err != nil {
		http.Error(w, "Failed to calculate secret hash", http.StatusInternalServerError)
		log.Printf("Failed to calculate secret hash: %v", err)
		return
	}
	obj, err := c.AuthClient.AdminInitiateAuth(
		context.TODO(),
		&cognitoidentityprovider.AdminInitiateAuthInput{
			AuthFlow:   "ADMIN_USER_PASSWORD_AUTH",
			ClientId:   &c.AuthClient.AppClientID,
			UserPoolId: &c.AuthClient.UserPoolID,
			AuthParameters: map[string]string{
				"USERNAME":    user.Username,
				"PASSWORD":    user.Password,
				"SECRET_HASH": secretHash,
			},
		},
	)
	if err != nil {
		http.Error(w, "Failed to sign in", http.StatusInternalServerError)
		log.Printf("Failed to sign in: %v", err)
		return
	}

	response := &SignInResponse{
		AccessToken:  obj.AuthenticationResult.AccessToken,
		ExpiresIn:    obj.AuthenticationResult.ExpiresIn,
		TokenType:    obj.AuthenticationResult.TokenType,
		IDToken:      obj.AuthenticationResult.IdToken,
		RefreshToken: obj.AuthenticationResult.RefreshToken,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to sign in", http.StatusInternalServerError)
		log.Printf("Failed to sign in: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jsonData))
}
