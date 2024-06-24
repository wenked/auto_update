package models

import "time"

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
