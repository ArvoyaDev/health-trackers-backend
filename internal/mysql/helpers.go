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

func (d *Database) GetUserBySub(cognitoSub string) (User, error) {
	query := `SELECT * FROM users WHERE cognito_sub = ?`
	row := d.mysql.QueryRow(query, cognitoSub)
	var user User
	err := row.Scan(&user.ID, &user.Email, &user.CognitoSub)
	if err != nil {
		return User{}, fmt.Errorf("error scanning user: %w", err)
	}
	return user, nil
}

func (d *Database) GetIllnessesByUserID(userID int) ([]Illness, error) {
	query := `SELECT * FROM illnesses WHERE user_id = ?`
	rows, err := d.mysql.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying illnesses: %w", err)
	}
	defer rows.Close()

	var illnesses []Illness
	for rows.Next() {
		var illness Illness
		err := rows.Scan(&illness.ID, &illness.IllnessName, &illness.UserID)
		if err != nil {
			return nil, fmt.Errorf("error scanning illness: %w", err)
		}
		illnesses = append(illnesses, illness)
	}
	return illnesses, nil
}

func (d *Database) GetSymptomsByIllnessID(illnessID int) ([]Symptom, error) {
	query := `SELECT * FROM symptoms WHERE illness_id = ?`
	rows, err := d.mysql.Query(query, illnessID)
	if err != nil {
		return nil, fmt.Errorf("error querying symptoms: %w", err)
	}
	defer rows.Close()

	var symptoms []Symptom
	for rows.Next() {
		var symptom Symptom
		err := rows.Scan(&symptom.ID, &symptom.SymptomName, &symptom.IllnessID)
		if err != nil {
			return nil, fmt.Errorf("error scanning symptom: %w", err)
		}
		symptoms = append(symptoms, symptom)
	}
	return symptoms, nil
}
