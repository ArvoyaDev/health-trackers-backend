package db

type User struct {
	ID         int    `json:"id"`
	CognitoSub string `json:"cognito_sub"`
	Email      string `json:"email"`
}

type Illness struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	IllnessName string `json:"illness_name"`
}

type Symptom struct {
	ID          int    `json:"id"`
	IllnessID   int    `json:"illness_id"`
	SymptomName string `json:"symptom_name"`
}

type SymptomLog struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	IllnessID int    `json:"illness_id"`
	LogTime   string `json:"log_time"`
	Severity  string `json:"severity"`
	Symptoms  string `json:"symptoms"`
	Notes     string `json:"notes"`
}

type CompleteUser struct {
	Email    string   `json:"email"`
	Illness  string   `json:"illness"`
	Symptoms []string `json:"symptoms"`
}
