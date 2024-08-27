package db

import (
	"fmt"
)

func (d *Database) CreateUser(user User) error {
	query := `INSERT INTO users (email, cognito_sub) VALUES (?, ?)`
	_, err := d.mysql.Exec(query, user.Email, user.CognitoSub)
	if err != nil {
		return fmt.Errorf("error inserting user: %w", err)
	}
	return nil
}

func (d *Database) CreateIllness(illness Illness) error {
	query := `INSERT INTO illnesses (illness_name, user_id) VALUES (?, ?)`
	_, err := d.mysql.Exec(query, illness.IllnessName, illness.UserID)
	if err != nil {
		return fmt.Errorf("error inserting illness: %w", err)
	}
	return nil
}

func (d *Database) CreateSymptom(symptom Symptom) error {
	query := `INSERT INTO symptoms (symptom_name, illness_id) VALUES (?, ?)`
	_, err := d.mysql.Exec(query, symptom.SymptomName, symptom.IllnessID)
	if err != nil {
		return fmt.Errorf("error inserting symptom: %w", err)
	}
	return nil
}

func (d *Database) CreateSymptomLog(symptomLog SymptomLog) error {
	query := `INSERT INTO logs (user_id, illness_id, log_time, severity, symptoms, notes) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := d.mysql.Exec(
		query,
		symptomLog.UserID,
		symptomLog.IllnessID,
		symptomLog.LogTime,
		symptomLog.Severity,
		symptomLog.Symptoms,
		symptomLog.Notes,
	)
	if err != nil {
		return fmt.Errorf("error inserting symptom log: %w", err)
	}
	return nil
}
