package models

import "database/sql"

type User struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Password  string `json:"password"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type CreateUser struct {
	Name          string `json:"name"`
	Password      string `json:"password"`
	Email         string `json:"email"`
	CreateUserKey string `json:"create_user_key"`
}

type UpdateUserRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
type UpdateUser struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func ScanUser(rows *sql.Rows) (User, error) {
	var n User
	err := rows.Scan(&n.ID, &n.Name, &n.Password, &n.Email, &n.CreatedAt, &n.UpdatedAt)
	return n, err
}

func ScanRowUser(row *sql.Row) (User, error) {
	var n User
	err := row.Scan(&n.ID, &n.Name, &n.Password, &n.Email, &n.CreatedAt, &n.UpdatedAt)
	return n, err
}
