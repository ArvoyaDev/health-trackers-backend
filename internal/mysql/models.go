package db

type User struct {
	ID         int    `json:"id"`
	CognitoSub string `json:"cognito_sub"`
	Email      string `json:"email"`
}

type Tracker struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	TrackerName string `json:"tracker_name"`
}

type Symptom struct {
	ID          int    `json:"id"`
	TrackerID   int    `json:"tracker_id"`
	SymptomName string `json:"symptom_name"`
}

type SymptomLog struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	TrackerID int    `json:"tracker_id"`
	LogTime   string `json:"log_time"`
	Severity  string `json:"severity"`
	Symptoms  string `json:"symptoms"`
	Notes     string `json:"notes"`
}

type CompleteUser struct {
	Email    string   `json:"email"`
	Tracker  string   `json:"tracker"`
	Symptoms []string `json:"symptoms"`
}

type SymptomLogRequestBody struct {
	TrackerName      string `json:"tracker_name"`
	SelectedSymptoms string `json:"selected_symptoms"`
	Severity         string `json:"severity"`
	Notes            string `json:"notes"`
	UserID           int    `json:"user_id"`
	TrackerID        int    `json:"tracker_id"`
}
