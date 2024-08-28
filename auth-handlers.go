package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ArvoyaDev/symptom-tracker-backend/internal/auth"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
)

type User struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
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
	err := cfg.AuthClient.SignUp(
		context.Background(),
		user.Username,
		user.FirstName,
		user.LastName,
		user.Password,
	)
	if err != nil {
		http.Error(w, "Failed to sign up user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
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
		return
	}

	w.WriteHeader(http.StatusOK)
}

type SignInResponse struct {
	AccessToken *string `json:"accessToken"`
	ExpiresIn   int32   `json:"expiresIn"`
	TokenType   *string `json:"tokenType"`
	IDToken     *string `json:"idToken"`
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
		error := "Failed to authenticate user: " + err.Error()
		http.Error(w, error, http.StatusInternalServerError)
		return
	}

	// get the sub value from the id token by decoding it
	// store the sub value in a cookie
	// Decode the JWT (ID token)
	idToken := *obj.AuthenticationResult.IdToken
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		http.Error(w, "Invalid ID token", http.StatusInternalServerError)
		return
	}

	// Decode the payload (the second part of the JWT)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		http.Error(w, "Failed to decode ID token", http.StatusInternalServerError)
		return
	}

	// Unmarshal the payload into a map to extract the "sub" claim
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		http.Error(w, "Failed to parse ID token", http.StatusInternalServerError)
		return
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		http.Error(w, "Failed to extract 'sub' from ID token", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    *obj.AuthenticationResult.RefreshToken,
		Path:     "/aws-cognito/refresh-token",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "userSub",
		Value:    sub, // Replace with the actual email value
		Path:     "/aws-cognito/refresh-token",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})

	response := &SignInResponse{
		AccessToken: obj.AuthenticationResult.AccessToken,
		ExpiresIn:   obj.AuthenticationResult.ExpiresIn,
		TokenType:   obj.AuthenticationResult.TokenType,
		IDToken:     obj.AuthenticationResult.IdToken,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jsonData))
}

func (c *config) RefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := r.Cookie("refreshToken")
	if err != nil {
		http.Error(w, "Failed to retrieve refresh token", http.StatusInternalServerError)
		return
	}
	userSub, err := r.Cookie("userSub")
	if err != nil {
		http.Error(w, "Failed to retrieve user email", http.StatusInternalServerError)
		return
	}

	secretHash, err := auth.CalculateSecretHash(
		c.AuthClient.AppClientID,
		os.Getenv("COGNITO_CLIENT_SECRET"),
		userSub.Value,
	)
	if err != nil {
		http.Error(w, "Failed to calculate secret hash", http.StatusInternalServerError)
		return
	}

	obj, err := c.AuthClient.AdminInitiateAuth(
		context.TODO(),
		&cognitoidentityprovider.AdminInitiateAuthInput{
			AuthFlow:   "REFRESH_TOKEN_AUTH",
			ClientId:   &c.AuthClient.AppClientID,
			UserPoolId: &c.AuthClient.UserPoolID,
			AuthParameters: map[string]string{
				"REFRESH_TOKEN": refreshToken.Value,
				"SECRET_HASH":   secretHash,
			},
		},
	)
	if err != nil {
		error := "Failed to refresh token: " + err.Error()
		http.Error(w, error, http.StatusInternalServerError)
		return
	}

	response := &SignInResponse{
		AccessToken: obj.AuthenticationResult.AccessToken,
		ExpiresIn:   obj.AuthenticationResult.ExpiresIn,
		TokenType:   obj.AuthenticationResult.TokenType,
		IDToken:     obj.AuthenticationResult.IdToken,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jsonData))
}
