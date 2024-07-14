package models

import (
	"database/sql"
	"time"
)

type UpdateServer struct {
	ID         int64     `json:"id"`
	Host       string    `json:"host"`
	Password   string    `json:"password"`
	Script     string    `json:"script"`
	PipelineID int64     `json:"pipeline_id"`
	Label      string    `json:"label"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func ScanUpdateServer(rows *sql.Rows) (UpdateServer, error) {
	var n UpdateServer
	err := rows.Scan(&n.ID, &n.Host, &n.Password, &n.Script, &n.PipelineID, &n.Label, &n.Active, &n.CreatedAt, &n.UpdatedAt)
	return n, err
}

func ScanRowUpdateServer(row *sql.Row) (UpdateServer, error) {
	var n UpdateServer
	err := row.Scan(&n.ID, &n.Host, &n.Password, &n.Script, &n.PipelineID, &n.Label, &n.Active, &n.CreatedAt, &n.UpdatedAt)
	return n, err
}
