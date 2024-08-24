package db

import (
	"fmt"
)

type HeartburnLog struct {
	ID       int    `json:"id"`
	LogTime  string `json:"log_time"`
	Severity string `json:"severity"`
	Symptoms string `json:"symptoms"`
	LastMeal string `json:"last_meal"`
	Password string `json:"password"`
}

func (d *Database) CreateHeartburnLog(log HeartburnLog) error {
	query := `INSERT INTO logs (severity, symptoms, last_meal, password) VALUES (?, ?, ?, ?)`
	_, err := d.mysql.Exec(query, log.Severity, log.Symptoms, log.LastMeal, log.Password)
	if err != nil {
		return fmt.Errorf("error inserting log: %w", err)
	}
	return nil
}

func (d *Database) GetHeartburnLogs() ([]HeartburnLog, error) {
	rows, err := d.mysql.Query(
		`SELECT id, log_time, severity, symptoms, last_meal FROM logs WHERE password = "sriswamisatchidananda" ORDER BY log_time DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("error querying logs: %w", err)
	}
	defer rows.Close()

	var logs []HeartburnLog
	for rows.Next() {
		var log HeartburnLog
		if err := rows.Scan(&log.ID, &log.LogTime, &log.Severity, &log.Symptoms, &log.LastMeal); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		logs = append(logs, log)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	return logs, nil
}
