package models

import "database/sql"

type Company struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func ScanCompany(rows *sql.Rows) (Company, error) {
	var n Company
	err := rows.Scan(&n.ID, &n.Name, &n.CreatedAt, &n.UpdatedAt)
	return n, err
}

func ScanRowCompany(row *sql.Row) (Company, error) {
	var n Company
	err := row.Scan(&n.ID, &n.Name, &n.CreatedAt, &n.UpdatedAt)
	return n, err
}
