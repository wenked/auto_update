package database

import (
	"auto-update/internal/database/models"
	"auto-update/utils"
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
	CreatePipeline(name string, id int64) (int64, error)
	UpdatePipeline(opts *models.UpdatePipeline, user_id int64) error
	DeletePipeline(id int64, user_id int64) error
	ListPipelines(user_id int64) ([]models.Pipeline, error)
	GetUserPipelineById(pipeline_id int64, user_id int64) (models.Pipeline, error)
	CreateUser(name string, email string, password string) (int64, error)
	UpdateUser(opts *models.User) error
	DeleteUser(id int64) error
	GetUserByEmail(email string) (models.User, error)
	GetUserByID(id int64) (models.User, error)
	ListUsers(page int64, limit int64) ([]models.User, error)
	CreateNotificationConfig(config *models.NotificationConfig) (int64, error)
	UpdateNotificationConfig(id int64, userId int64, notificationConfig *models.NotificationConfig) error
	DeleteNotificationConfig(id int64, userId int64) error
	GetUserNotificationConfig(id int64, userId int64) (models.NotificationConfig, error)
	GetUserNotificationByType(userId int64, notificationType string) ([]models.NotificationConfig, error)
	UpdateServersPasswords() error
}

type ScanFunc[T any] func(*sql.Rows) (T, error)

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

func ScanRows[T any](rows *sql.Rows, scanFunc ScanFunc[T]) ([]T, error) {
	var items []T

	for rows.Next() {
		item, err := scanFunc(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
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

	server, err := models.ScanRowUpdateServer(row)

	if err != nil {
		slog.Error("error in user server query", "error", err)
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
	fmt.Println("pipeline_id", pipeline_id)
	rows, err := s.db.QueryContext(ctx, `SELECT * FROM "servers" WHERE pipeline_id = ? AND active = 1`, pipeline_id)

	if err != nil {
		fmt.Println("error in query", err)
		return nil, err
	}

	defer rows.Close()

	servers, err := ScanRows(rows, models.ScanUpdateServer)

	if err != nil {
		slog.Error("error scaning servers rows", "error", err)
		return nil, err
	}

	fmt.Println("servers", servers)

	return servers, nil
}

func (s *service) CreatePipeline(name string, user_id int64) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.db.ExecContext(ctx, `INSERT INTO pipelines (name,user_id) VALUES (?,?)`, name, user_id)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *service) UpdatePipeline(opts *models.UpdatePipeline, user_id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if opts.Name != "" {
		_, err := s.db.ExecContext(ctx, `UPDATE pipelines SET Name = ? WHERE id = ? and user_id = ?`, opts.Name, opts.ID, user_id)
		if err != nil {
			fmt.Println("error in update name", err)
			return err
		}
	}

	return nil
}

func (s *service) DeletePipeline(id int64, user_id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx, `DELETE FROM pipelines WHERE id = ? and user_id = ?`, id, user_id)
	if err != nil {
		fmt.Println("error in delete", err)
		return err
	}

	return nil
}

func (s *service) ListPipelines(user_id int64) ([]models.Pipeline, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, `SELECT * FROM pipelines WHERE user_id = ?`, user_id)

	if err != nil {
		fmt.Println("error in query", err)
		return nil, err
	}

	defer rows.Close()

	pipelines, err := ScanRows(rows, models.ScanPipeline)

	if err != nil {
		slog.Error("error scaning pipeline rows", "error", err)
		return nil, err
	}

	return pipelines, nil
}

func (s *service) GetUserPipelineById(pipeline_id int64, user_id int64) (models.Pipeline, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	row := s.db.QueryRowContext(ctx, `SELECT * FROM pipelines WHERE id = ? and user_id = ?`, pipeline_id, user_id)

	pipeline, err := models.ScanRowPipeline(row)

	if err != nil {
		slog.Error("error in user pipeline query", "error", err)
		return models.Pipeline{}, err
	}

	return pipeline, nil

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
		slog.Error("error in delete", "error", err)
		return err
	}

	return nil
}

func (s *service) GetUserByEmail(email string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	row := s.db.QueryRowContext(ctx, `SELECT * FROM users WHERE email = ?`, email)

	user, err := models.ScanRowUser(row)
	if err != nil {
		slog.Error("error in select user query", "error", err)
		return models.User{}, err
	}

	return user, nil
}

func (s *service) GetUserByID(id int64) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	row := s.db.QueryRowContext(ctx, `SELECT * FROM users WHERE id = ?`, id)

	user, err := models.ScanRowUser(row)
	if err != nil {
		slog.Error("error in select user query", "error", err)
		return models.User{}, err
	}

	return user, nil
}

func (s *service) ListUsers(page int64, limit int64) ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	offset := (page - 1) * limit
	rows, err := s.db.QueryContext(ctx, `SELECT * FROM users LIMIT ? OFFSET ?`, page, offset)
	var users []models.User
	if err != nil {
		slog.Error("error in select users query", "error", err)
		return users, err
	}

	defer rows.Close()

	users, err = ScanRows(rows, models.ScanUser)

	if err != nil {
		slog.Error("error scaning users", "error", nil)
		return users, err
	}

	return users, nil
}

func (s *service) CreateNotificationConfig(config *models.NotificationConfig) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.db.ExecContext(ctx, `INSERT INTO notification_config (type,name,number,user_id,url) VALUES (?,?,?,?,?)`, config.Type, config.Name, config.Number, config.UserID, config.Url)
	if err != nil {
		slog.Error("error inserting notification config", "error", err)
		return 0, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		slog.Error("error getting notification config id", "error", err)
		return 0, err
	}

	return id, nil
}

func (s *service) UpdateNotificationConfig(id int64, userId int64, config *models.NotificationConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if config.Name != "" {
		_, err := s.db.ExecContext(ctx, `UPDATE  notification_config SET name = ? WHERE id = ? and user_id = ?`, config.Name, id, userId)
		if err != nil {
			slog.Error("error update notification config id", "error", err)
			return err
		}
	}

	if config.Number != "" {
		_, err := s.db.ExecContext(ctx, `UPDATE  notification_config SET number = ? WHERE id = ? and user_id = ?`, config.Number, id, userId)
		if err != nil {
			slog.Error("error update notification config id", "error", err)
			return err
		}
	}

	if config.Type != "" {
		_, err := s.db.ExecContext(ctx, `UPDATE notification_config SET type = ? WHERE id = ? and user_id = ?`, config.Type, id, userId)
		if err != nil {
			slog.Error("error update notification config id", "error", err)
			return err
		}
	}

	if config.Url != "" {
		_, err := s.db.ExecContext(ctx, `UPDATE notification_config SET url = ? WHERE id = ? and user_id = ?`, config.Url, id, userId)
		if err != nil {
			slog.Error("error update notification config id", "error", err)
			return err
		}
	}

	return nil
}

func (s *service) DeleteNotificationConfig(id int64, userId int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx, `DELETE from notification_config WHERE id = ? and user_id = ? `, id, userId)

	if err != nil {
		slog.Error("Error deleting notification config", "error", err)
		return err
	}

	return nil
}

func (s *service) GetUserNotificationConfig(id int64, userId int64) (models.NotificationConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	row := s.db.QueryRowContext(ctx, `SELECT * from notification_config where id = ? and user_id = ?`, id, userId)

	var notificationConfig models.NotificationConfig

	notificationConfig, err := models.ScanRowNotificationConfig(row)

	if err != nil {
		slog.Error("error in GetUserNotificationConfig query", "error", err)
		return models.NotificationConfig{}, err
	}

	return notificationConfig, nil
}

func (s *service) GetUserNotificationByType(userId int64, notificationType string) ([]models.NotificationConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var notificationConfigs []models.NotificationConfig
	rows, err := s.db.QueryContext(ctx, `SELECT * FROM notification_config WHERE user_id = ? and type = ?`, userId, notificationType)

	if err != nil {
		slog.Error("error in GetUserNotificationByType query", "error", err)
		return notificationConfigs, err
	}

	defer rows.Close()

	notificationConfigs, err = ScanRows(rows, models.ScanNotificationConfig)

	if err != nil {
		slog.Error("error scaning rows", "error", err)
		return notificationConfigs, err
	}

	return notificationConfigs, nil
}

func (s *service) UpdateServersPasswords() error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, `SELECT id,password FROM servers`)

	if err != nil {
		slog.Error("error in query", "err", err)
		return err
	}

	defer rows.Close()

	// var servers []models.UpdateServer
	for rows.Next() {
		var server models.UpdateServer
		err := rows.Scan(&server.ID, &server.Password)
		//	err := rows.Scan(&server.ID, &server.Host, &server.Password, &server.Script, &server.PipelineID, &server.Label, &server.Active, &server.CreatedAt, &server.UpdatedAt)
		if err != nil {
			slog.Error("error scaning rows", "error", err)
			return err
		}

		hashedPassowrd, err := utils.Encrypt(server.Password)
		if err != nil {
			slog.Error("error hashing password", "error", err)
			continue
		}
		fmt.Println("server.Password", server.Password)
		fmt.Println("hashedpassword", hashedPassowrd)
		fmt.Println("ID", server.ID)
		_, err = s.db.ExecContext(ctx, "UPDATE servers SET password = ? WHERE id = ?", hashedPassowrd, server.ID)

		if err != nil {
			slog.Error("error saving in db", "error", err)
		}
		//	servers = append(servers, server)
	}

	//	fmt.Println("servers", servers)

	return nil
}
