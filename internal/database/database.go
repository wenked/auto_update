package database

import (
	"auto-update/internal/database/models"
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type Service interface {
	Health() map[string]string
	CreateUpdate(pusher_name string, branch string, status string, message string) (int64, error)
	UpdateStatusAndMessage(id int64, status string, message string) error
	GetUpdates(limit int, offset int) ([]Update, error)
	CreateServer(host string, password string, script string, pipeline_id int64, label string) (int64, error)
	UpdateServer(opts *models.UpdateServer) error
	GetServer(id int64) (*models.UpdateServer, error)
	DeleteServer(id int64) error
	ListServers(pipeline_id int64) ([]models.UpdateServer, error)
	CreatePipeline(name string) (int64, error)
	UpdatePipeline(opts *models.UpdatePipeline) error
	DeletePipeline(id int64) error
	ListPipelines() ([]models.UpdatePipeline, error)
	CreateUser(name string, email string, password string) (int64, error)
	UpdateUser(opts *models.User) error
	DeleteUser(id int64) error
	GetUserByEmail(email string) (models.User, error)
	GetUserByID(id int64) (models.User, error)
	ListUsers() ([]models.User, error)
}

type service struct {
	db *sql.DB
}

type Update struct {
	ID         int64     `json:"id"`
	PusherName string    `json:"pusher_name"`
	Branch     string    `json:"branch"`
	Status     string    `json:"status"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

var (
	dburl = os.Getenv("DB_URL")
)

func New() Service {
	db, err := sql.Open("libsql", dburl)
	if err != nil {
		slog.Error("error connecting to database", err)
		log.Fatal(err)
	}

	_, migration_err := db.Exec(`PRAGMA foreign_keys = ON;`)
	if migration_err != nil {
		log.Fatal(migration_err)
	}

	s := &service{db: db}
	return s
}

var s = New()

func GetService() Service {
	return s
}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := s.db.PingContext(ctx)
	if err != nil {
		log.Fatalf(fmt.Sprintf("db down: %v", err))
	}

	return map[string]string{
		"message": "It's healthy",
	}
}

func (s *service) CreateUpdate(pusher_name string, branch string, status string, message string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.db.ExecContext(ctx, `INSERT INTO updates (pusher_name, branch, status, message) VALUES (?, ?, ?, ?)`, pusher_name, branch, status, message)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *service) UpdateStatusAndMessage(id int64, status string, message string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx, `UPDATE updates SET status = ?,message = ? WHERE id = ?`, status, message, id)
	if err != nil {
		fmt.Println("error in update status and message", err)
		return err
	}

	return nil
}

func (s *service) GetUpdates(limit int, offset int) ([]Update, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, `SELECT * FROM updates ORDER BY id DESC, created_at DESC LIMIT ? OFFSET ?`, limit, offset)

	if err != nil {
		fmt.Println("error in query", err)
		return nil, err
	}

	defer rows.Close()

	var updates []Update
	for rows.Next() {
		var update Update
		err := rows.Scan(&update.ID, &update.PusherName, &update.Branch, &update.Status, &update.Message, &update.CreatedAt, &update.UpdatedAt)
		if err != nil {
			fmt.Println("error", err)
			return nil, err
		}

		updates = append(updates, update)
	}

	return updates, nil
}

func (s *service) CreateServer(host string, password string, script string, pipeline_id int64, label string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.db.ExecContext(ctx, `INSERT INTO servers (host, password, script,pipeline_id,label) VALUES (?, ?, ? , ?, ?)`, host, password, script, pipeline_id, label)
	if err != nil {
		fmt.Println("error in insert", err)
		return 0, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *service) UpdateServer(opts *models.UpdateServer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if opts.Host != "" {
		_, err := s.db.ExecContext(ctx, `UPDATE servers SET Host = ? WHERE id = ?`, opts.Host, opts.ID)
		if err != nil {
			fmt.Println("error in update host", err)
			return err
		}
	}

	if opts.Password != "" {
		_, err := s.db.ExecContext(ctx, `UPDATE servers SET Password = ? WHERE id = ?`, opts.Password, opts.ID)
		if err != nil {
			fmt.Println("error in update password", err)
			return err
		}
	}

	if opts.Script != "" {
		_, err := s.db.ExecContext(ctx, `UPDATE servers SET script = ? WHERE id = ?`, opts.Script, opts.ID)
		if err != nil {
			fmt.Println("error in update script", err)
			return err
		}
	}

	if opts.Label != "" {
		_, err := s.db.ExecContext(ctx, `UPDATE servers SET label = ? WHERE id = ?`, opts.Label, opts.ID)
		if err != nil {
			fmt.Println("error in update label", err)

			return err
		}
	}

	if opts.Active {
		_, err := s.db.ExecContext(ctx, `UPDATE servers SET active = ? WHERE id = ?`, opts.Active, opts.ID)
		if err != nil {
			fmt.Println("error in update active", err)
			return err
		}
	}

	return nil
}

func (s *service) GetServer(id int64) (*models.UpdateServer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	row := s.db.QueryRowContext(ctx, `SELECT * FROM servers WHERE id = ?`, id)

	var server models.UpdateServer
	err := row.Scan(&server.ID, &server.Host, &server.Password, &server.Script, &server.PipelineID, &server.CreatedAt, &server.UpdatedAt, &server.Label, &server.Active)
	if err != nil {
		fmt.Println("error in query", err)
		return nil, err
	}

	return &server, nil
}

func (s *service) DeleteServer(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx, `DELETE FROM servers WHERE id = ?`, id)
	if err != nil {
		fmt.Println("error in delete", err)
		return err
	}

	return nil
}

func (s *service) ListServers(pipeline_id int64) ([]models.UpdateServer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, `SELECT * FROM servers WHERE pipeline_id = ? AND active = 1`, pipeline_id)

	if err != nil {
		fmt.Println("error in query", err)
		return nil, err
	}

	defer rows.Close()

	var servers []models.UpdateServer
	for rows.Next() {
		var server models.UpdateServer
		err := rows.Scan(&server.ID, &server.Host, &server.Password, &server.Script, &server.PipelineID, &server.CreatedAt, &server.UpdatedAt, &server.Label, &server.Active)
		if err != nil {
			fmt.Println("error", err)
			return nil, err
		}

		servers = append(servers, server)
	}

	return servers, nil
}

func (s *service) CreatePipeline(name string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.db.ExecContext(ctx, `INSERT INTO pipelines (name) VALUES (?)`, name)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *service) UpdatePipeline(opts *models.UpdatePipeline) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if opts.Name != "" {
		_, err := s.db.ExecContext(ctx, `UPDATE pipelines SET Name = ? WHERE id = ?`, opts.Name, opts.ID)
		if err != nil {
			fmt.Println("error in update name", err)
			return err
		}
	}

	return nil
}

func (s *service) DeletePipeline(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx, `DELETE FROM pipelines WHERE id = ?`, id)
	if err != nil {
		fmt.Println("error in delete", err)
		return err
	}

	return nil
}

func (s *service) ListPipelines() ([]models.UpdatePipeline, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, `SELECT * FROM pipelines`)

	if err != nil {
		fmt.Println("error in query", err)
		return nil, err
	}

	defer rows.Close()

	var pipelines []models.UpdatePipeline
	for rows.Next() {
		var pipeline models.UpdatePipeline
		err := rows.Scan(&pipeline.ID, &pipeline.Name)
		if err != nil {
			fmt.Println("error", err)
			return nil, err
		}

		pipelines = append(pipelines, pipeline)
	}

	return pipelines, nil
}

func (s *service) CreateUser(name string, email string, password string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.db.ExecContext(ctx, `INSERT INTO users (name,email, password) VALUES (?, ?, ?)`, name, email, password)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *service) UpdateUser(opts *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if opts.Name != "" {
		_, err := s.db.ExecContext(ctx, `UPDATE users SET name = ? WHERE id = ?`, opts.Name, opts.ID)
		if err != nil {
			slog.Error("error in update name", err)
			return err
		}
	}

	if opts.Email != "" {
		_, err := s.db.ExecContext(ctx, `UPDATE users SET email = ? WHERE id = ?`, opts.Email, opts.ID)
		if err != nil {
			slog.Error("error in update email", err)
			return err
		}
	}

	if opts.Password != "" {
		_, err := s.db.ExecContext(ctx, `UPDATE users SET password = ? WHERE id = ?`, opts.Password, opts.ID)
		if err != nil {
			slog.Error("error in update password", err)
			return err
		}
	}

	return nil
}

func (s *service) DeleteUser(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		slog.Error("error in delete", err)
		return err
	}

	return nil
}

func (s *service) GetUserByEmail(email string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	row := s.db.QueryRowContext(ctx, `SELECT * FROM users WHERE email = ?`, email)

	var user models.User
	err := row.Scan(&user.ID, &user.Name, &user.Password, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		slog.Error("error in query", err)
		return models.User{}, err
	}

	return user, nil
}

func (s *service) GetUserByID(id int64) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	row := s.db.QueryRowContext(ctx, `SELECT * FROM users WHERE id = ?`, id)

	var user models.User
	err := row.Scan(&user.ID, &user.Name, &user.Password, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		slog.Error("error in query", err)
		return models.User{}, err
	}

	return user, nil
}

func (s *service) ListUsers() ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, `SELECT * FROM users`)

	if err != nil {
		slog.Error("error in query", err)
		return nil, err
	}

	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Name, &user.Password, &user.Email, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			slog.Error("error", err)
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}
