package models

import "database/sql"

type Pipeline struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	UserID    int64  `json:"user_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type UpdatePipeline struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func ScanPipeline(rows *sql.Rows) (Pipeline, error) {
	var n Pipeline
	err := rows.Scan(&n.ID, &n.Name, &n.UserID, &n.CreatedAt, &n.UpdatedAt)
	return n, err
}

func ScanRowPipeline(row *sql.Row) (Pipeline, error) {
	var n Pipeline
	err := row.Scan(&n.ID, &n.Name, &n.UserID, &n.CreatedAt, &n.UpdatedAt)
	return n, err
}
