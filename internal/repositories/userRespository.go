package repository

import (
	"auto-update/internal/database"
	"auto-update/internal/database/models"
	"database/sql"
	"fmt"
	"log/slog"

	"golang.org/x/net/context"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user models.User) (int64, error)
	UpdateUser(ctx context.Context, user models.User) (int64, error)
	DeleteUser(ctx context.Context, id int64) error
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	GetUserByID(ctx context.Context, id int64) (models.User, error)
	ListUsers(ctx context.Context, page int64, limit int64) ([]models.User, error)
}

type SQLUseryRepository struct {
	db *sql.DB
}

func NewUseryRepository(db *sql.DB) *SQLUseryRepository {
	return &SQLUseryRepository{
		db: db,
	}
}

func (s *SQLUseryRepository) CreateUser(ctx context.Context, user models.User) (int64, error) {

	result, err := s.db.ExecContext(ctx, `INSERT INTO users (name, email, password, company_id) VALUES ($1, $2, $3, $4)`, user.Name, user.Email, user.Password, user.CompanyID)
	fmt.Println("bolamas")
	if err != nil {
		fmt.Println(err)
		slog.Error("error creating user", "error", err)
		return 0, err
	}

	fmt.Println("bolamas2")
	id, err := result.LastInsertId()

	if err != nil {
		fmt.Println(err)
		slog.Error("error creating user", "error", err)
		return 0, err
	}

	return id, nil
}

func (s *SQLUseryRepository) UpdateUser(ctx context.Context, user models.User) error {

	if user.Name != "" {
		_, err := s.db.ExecContext(ctx, `UPDATE users SET name = ? WHERE id = ?`, user.Name, user.ID)
		if err != nil {
			slog.Error("error in update name", "error", err)
			return err
		}
	}

	if user.Email != "" {
		_, err := s.db.ExecContext(ctx, `UPDATE users SET email = ? WHERE id = ?`, user.Email, user.ID)
		if err != nil {
			slog.Error("error in update email", "error", err)
			return err
		}
	}

	if user.Password != "" {
		_, err := s.db.ExecContext(ctx, `UPDATE users SET password = ? WHERE id = ?`, user.Password, user.ID)
		if err != nil {
			slog.Error("error in update password", "error", err)
			return err
		}
	}

	return nil
}

func (s *SQLUseryRepository) DeleteUser(ctx context.Context, id int64) error {

	_, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		slog.Error("error in delete", "error", err)
		return err
	}

	return nil
}

func (s *SQLUseryRepository) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	row := s.db.QueryRowContext(ctx, `SELECT * FROM users WHERE email = ?`, email)

	user, err := models.ScanRowUser(row)
	if err != nil {
		slog.Error("error in select user query", "error", err)
		return models.User{}, err
	}

	return user, nil
}

func (s *SQLUseryRepository) GetUserByID(ctx context.Context, id int64) (models.User, error) {

	row := s.db.QueryRowContext(ctx, `SELECT * FROM users WHERE id = ?`, id)

	user, err := models.ScanRowUser(row)
	if err != nil {
		slog.Error("error in select user query", "error", err)
		return models.User{
			ID: 0,
		}, err
	}

	return user, nil
}

func (s *SQLUseryRepository) ListUsers(ctx context.Context, page int64, limit int64) ([]models.User, error) {
	offset := (page - 1) * limit
	rows, err := s.db.QueryContext(ctx, `SELECT * FROM users LIMIT ? OFFSET ?`, page, offset)
	var users []models.User
	if err != nil {
		slog.Error("error in select users query", "error", err)
		return users, err
	}

	defer rows.Close()

	users, err = database.ScanRows(rows, models.ScanUser)

	if err != nil {
		slog.Error("error scaning users", "error", nil)
		return users, err
	}

	return users, nil
}
