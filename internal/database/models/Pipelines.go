package models

import (
	"database/sql"
	"time"
)

type Pipeline struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdatePipeline struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func ScanPipeline(rows *sql.Rows) (Pipeline, error) {
	var n Pipeline
	err := rows.Scan(&n.ID, &n.Name, &n.CreatedAt, &n.UpdatedAt, &n.UserID)
	return n, err
}

func ScanRowPipeline(row *sql.Row) (Pipeline, error) {
	var n Pipeline
	err := row.Scan(&n.ID, &n.Name, &n.CreatedAt, &n.UpdatedAt, &n.UserID)
	return n, err
}
