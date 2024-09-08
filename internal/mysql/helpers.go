package db

import (
	"fmt"
)

func (d *Database) CreateUser(email, sub string) error {
	query := `INSERT INTO users (email, cognito_sub) VALUES (?, ?)`
	_, err := d.mysql.Exec(query, email, sub)
	if err != nil {
		return fmt.Errorf("error inserting user: %w", err)
	}
	return nil
}

func (d *Database) CreateTracker(tracker string, userID int) error {
	query := `INSERT INTO trackers (tracker_name, user_id) VALUES (?, ?)`
	_, err := d.mysql.Exec(query, tracker, userID)
	if err != nil {
		return fmt.Errorf("error inserting tracker: %w", err)
	}
	return nil
}

func (d *Database) CreateSymptom(symptom string, trackerID int) error {
	query := `INSERT INTO symptoms (symptom_name, tracker_id) VALUES (?, ?)`
	_, err := d.mysql.Exec(query, symptom, trackerID)
	if err != nil {
		return fmt.Errorf("error inserting symptom: %w", err)
	}
	return nil
}

func (d *Database) CreateSymptomLog(symptomLog SymptomLogRequestBody) error {
	query := `INSERT INTO symptom_logs (user_id, tracker_id, severity, symptoms, notes) VALUES (?, ?, ?, ?, ?)`
	_, err := d.mysql.Exec(
		query,
		symptomLog.UserID,
		symptomLog.TrackerID,
		symptomLog.Severity,
		symptomLog.SelectedSymptoms,
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
	err := row.Scan(&user.ID, &user.CognitoSub, &user.Email)
	if err != nil {
		return User{}, fmt.Errorf("error scanning user: %w", err)
	}
	return user, nil
}

func (d *Database) GetTrackerByNameAndUserID(trackerName string, userID int) (Tracker, error) {
	query := `SELECT * FROM trackers WHERE tracker_name= ? AND user_id= ?`
	row := d.mysql.QueryRow(query, trackerName, userID)
	var tracker Tracker
	err := row.Scan(&tracker.ID, &tracker.UserID, &tracker.TrackerName)
	if err != nil {
		return Tracker{}, fmt.Errorf("error scanning tracker: %w", err)
	}
	return tracker, nil
}

func (d *Database) GetTrackerByUserID(userID int) ([]Tracker, error) {
	query := `SELECT * FROM trackers WHERE user_id = ?`
	rows, err := d.mysql.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying trackers: %w", err)
	}
	defer rows.Close()

	var trackers []Tracker
	for rows.Next() {
		var tracker Tracker
		err := rows.Scan(&tracker.ID, &tracker.UserID, &tracker.TrackerName)
		if err != nil {
			return nil, fmt.Errorf("error scanning tracker: %w", err)
		}
		trackers = append(trackers, tracker)
	}
	return trackers, nil
}

func (d *Database) GetSymptomsByTrackerID(trackerID int) ([]Symptom, error) {
	query := `SELECT * FROM symptoms WHERE tracker_id= ?`
	rows, err := d.mysql.Query(query, trackerID)
	if err != nil {
		return nil, fmt.Errorf("error querying symptoms: %w", err)
	}
	defer rows.Close()

	var symptoms []Symptom
	for rows.Next() {
		var symptom Symptom
		err := rows.Scan(&symptom.ID, &symptom.TrackerID, &symptom.SymptomName)
		if err != nil {
			return nil, fmt.Errorf("error scanning symptom: %w", err)
		}
		symptoms = append(symptoms, symptom)
	}
	return symptoms, nil
}

func (d *Database) GetSymptomLogsByUserID(userID int) ([]SymptomLog, error) {
	query := `SELECT * FROM symptom_logs WHERE user_id= ?`
	rows, err := d.mysql.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying symptom logs: %w", err)
	}
	defer rows.Close()

	var symptomLogs []SymptomLog
	for rows.Next() {
		var symptomLog SymptomLog
		err := rows.Scan(
			&symptomLog.ID,
			&symptomLog.UserID,
			&symptomLog.TrackerID,
			&symptomLog.LogTime,
			&symptomLog.Severity,
			&symptomLog.Symptoms,
			&symptomLog.Notes,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning symptom log: %w", err)
		}
		symptomLogs = append(symptomLogs, symptomLog)
	}
	return symptomLogs, nil
}
