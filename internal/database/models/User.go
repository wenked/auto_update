package models

type User struct {
	ID        int64  `json:"id"`
	Name      string `json:"username"`
	Password  string `json:"password"`
	Email     string `json:"email"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type CreateUser struct {
	Name          string `json:"username"`
	Password      string `json:"password"`
	Email         string `json:"email"`
	CreateUserKey string `json:"create_user_key"`
}
