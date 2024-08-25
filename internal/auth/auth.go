package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	cip "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

type CognitoClient struct {
	AppClientID string
	UserPoolID  string
	*cip.Client
}

// Init initializes the Cognito client
func Init() *CognitoClient {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	return &CognitoClient{
		AppClientID: os.Getenv("COGNITO_APP_CLIENT_ID"),
		Client:      cip.NewFromConfig(cfg),
		UserPoolID:  os.Getenv("COGNITO_USER_POOL_ID"),
	}
}

// SignUp registers a new user in the Cognito User Pool
func (c *CognitoClient) SignUp(ctx context.Context, email, password string) error {
	secretHash, err := CalculateSecretHash(c.AppClientID, os.Getenv("COGNITO_CLIENT_SECRET"), email)
	if err != nil {
		return errors.New("failed to calculate secret hash: " + err.Error())
	}
	input := &cip.SignUpInput{
		ClientId:   aws.String(c.AppClientID),
		Username:   aws.String(email),
		Password:   aws.String(password),
		SecretHash: aws.String(secretHash),
		UserAttributes: []types.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(email),
			},
		},
	}

	// Call the Cognito SignUp API
	_, err = c.Client.SignUp(ctx, input)
	if err != nil {
		return errors.New("failed to sign up user: " + err.Error())
	}

	return nil
}

func CalculateSecretHash(clientID, clientSecret, username string) (string, error) {
	key := []byte(clientSecret)
	msg := username + clientID

	h := hmac.New(sha256.New, key)
	_, err := h.Write([]byte(msg))
	if err != nil {
		return "", fmt.Errorf("failed to write to hash: %w", err)
	}
	hash := h.Sum(nil)

	// Encode to Base64
	secretHash := base64.StdEncoding.EncodeToString(hash)
	return secretHash, nil
}

type ConfirmSignupRequest struct {
	Username         string `json:"username"`
	ConfirmationCode string `json:"confirmationCode"`
}

func (c *CognitoClient) ConfirmSignup(w http.ResponseWriter, r *http.Request) {
	var req ConfirmSignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	secretHash, err := CalculateSecretHash(
		c.AppClientID,
		os.Getenv("COGNITO_CLIENT_SECRET"),
		req.Username,
	)
	if err != nil {
		http.Error(w, "Failed to calculate secret hash", http.StatusInternalServerError)
		log.Printf("Failed to calculate secret hash: %v", err)
		return
	}

	_, err = c.Client.ConfirmSignUp(context.TODO(), &cip.ConfirmSignUpInput{
		ClientId:         aws.String(c.AppClientID),
		Username:         aws.String(req.Username),
		SecretHash:       aws.String(secretHash),
		ConfirmationCode: aws.String(req.ConfirmationCode),
	})
	if err != nil {
		http.Error(w, "Failed to confirm signup", http.StatusInternalServerError)
		log.Printf("Failed to confirm signup: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
