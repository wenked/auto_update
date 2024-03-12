package database

import (
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
	UpdateServer(opts *UpdateServer) error
	GetServer(id int64) (*UpdateServer, error)
	DeleteServer(id int64) error
	ListServers(pipeline_id int64) ([]UpdateServer, error)
	CreatePipeline(name string) (int64, error)
	UpdatePipeline(opts *UpdatePipeline) error
	DeletePipeline(id int64) error
	ListPipelines() ([]UpdatePipeline, error)
}

type service struct {
	db *sql.DB
}

type UpdateServer struct {
	ID         int64     `json:"id"`
	Host       string    `json:"host"`
	Password   string    `json:"password"`
	Script     string    `json:"script"`
	PipelineID int64     `json:"pipeline_id"`
	Label      string    `json:"label"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type UpdatePipeline struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
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
		error := fmt.Sprintf("migration_err: %v", migration_err)
		fmt.Println(error)
		log.Fatal(migration_err)
	}

	_, migration_err = db.Exec(`CREATE TABLE IF NOT EXISTS "updates" (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			"pusher_name" TEXT,
			"branch" TEXT,
			"status" TEXT,
			"message" TEXT,
			"created_at" DATETIME DEFAULT CURRENT_TIMESTAMP,
			"updated_at" DATETIME DEFAULT CURRENT_TIMESTAMP
		)`)

	if migration_err != nil {
		error := fmt.Sprintf("migration_err: %v", migration_err)
		fmt.Println(error)
		log.Fatal(migration_err)
	}

	_, migration_err = db.Exec(`CREATE TABLE IF NOT EXISTS pipelines (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)

	if migration_err != nil {
		error := fmt.Sprintf("migration_err: %v", migration_err)
		fmt.Println(error)
		log.Fatal(migration_err)

	}

	_, migration_err = db.Exec(`CREATE TABLE IF NOT EXISTS servidores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		host TEXT,
		password TEXT,
		script TEXT,
		pipeline_id,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (pipeline_id) REFERENCES pipelines (id) ON DELETE CASCADE
	);`)

	if migration_err != nil {
		error := fmt.Sprintf("migration_err: %v", migration_err)
		fmt.Println(error)
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

	result, err := s.db.ExecContext(ctx, `INSERT INTO servidores (host, password, script,pipeline_id,label) VALUES (?, ?, ? , ?, ?)`, host, password, script, pipeline_id, label)
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

func (s *service) UpdateServer(opts *UpdateServer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if opts.Host != "" {
		_, err := s.db.ExecContext(ctx, `UPDATE servidores SET Host = ? WHERE id = ?`, opts.Host, opts.ID)
		if err != nil {
			fmt.Println("error in update host", err)
			return err
		}
	}

	if opts.Password != "" {
		_, err := s.db.ExecContext(ctx, `UPDATE servidores SET Password = ? WHERE id = ?`, opts.Password, opts.ID)
		if err != nil {
			fmt.Println("error in update password", err)
			return err
		}
	}

	if opts.Script != "" {
		_, err := s.db.ExecContext(ctx, `UPDATE servidores SET script = ? WHERE id = ?`, opts.Script, opts.ID)
		if err != nil {
			fmt.Println("error in update script", err)
			return err
		}
	}

	if opts.Label != "" {
		_, err := s.db.ExecContext(ctx, `UPDATE servidores SET label = ? WHERE id = ?`, opts.Label, opts.ID)
		if err != nil {
			fmt.Println("error in update label", err)

			return err
		}
	}

	return nil
}

func (s *service) GetServer(id int64) (*UpdateServer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	row := s.db.QueryRowContext(ctx, `SELECT * FROM servidores WHERE id = ?`, id)

	var server UpdateServer
	err := row.Scan(&server.ID, &server.Host, &server.Password, &server.Script)
	if err != nil {
		fmt.Println("error in query", err)
		return nil, err
	}

	return &server, nil
}

func (s *service) DeleteServer(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx, `DELETE FROM servidores WHERE id = ?`, id)
	if err != nil {
		fmt.Println("error in delete", err)
		return err
	}

	return nil
}

func (s *service) ListServers(pipeline_id int64) ([]UpdateServer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, `SELECT * FROM servidores WHERE pipeline_id = ?`, pipeline_id)

	if err != nil {
		fmt.Println("error in query", err)
		return nil, err
	}

	defer rows.Close()

	var servers []UpdateServer
	for rows.Next() {
		var server UpdateServer
		err := rows.Scan(&server.ID, &server.Host, &server.Password, &server.Script, &server.PipelineID, &server.CreatedAt, &server.UpdatedAt, &server.Label)
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

func (s *service) UpdatePipeline(opts *UpdatePipeline) error {
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

func (s *service) ListPipelines() ([]UpdatePipeline, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, `SELECT * FROM pipelines`)

	if err != nil {
		fmt.Println("error in query", err)
		return nil, err
	}

	defer rows.Close()

	var pipelines []UpdatePipeline
	for rows.Next() {
		var pipeline UpdatePipeline
		err := rows.Scan(&pipeline.ID, &pipeline.Name)
		if err != nil {
			fmt.Println("error", err)
			return nil, err
		}

		pipelines = append(pipelines, pipeline)
	}

	return pipelines, nil
}
