package models

import "database/sql"

type NotificationConfig struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Number    string `json:"number"`
	UserID    int64  `json:"user_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Url       string `json:"url"`
}

func ScanNotificationConfig(rows *sql.Rows) (NotificationConfig, error) {
	var n NotificationConfig
	err := rows.Scan(&n.ID, &n.Type, &n.Name, &n.Number, &n.UserID, &n.CreatedAt, &n.UpdatedAt, &n.Url)
	return n, err
}

func ScanRowNotificationConfig(row *sql.Row) (NotificationConfig, error) {
	var n NotificationConfig
	err := row.Scan(&n.ID, &n.Type, &n.Name, &n.Number, &n.UserID, &n.CreatedAt, &n.UpdatedAt, &n.Url)
	return n, err
}
