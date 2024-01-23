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
	UpdateStatusAndMessage(id int64, status string,message string) error
	GetUpdates(limit int ,offset int) ([]Update, error)
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
		slog.Error("error connecting to database",err)
		log.Fatal(err)
	}

	

	
	_,migration_err := db.Exec(`CREATE TABLE IF NOT EXISTS "updates" (
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

func (s *service) CreateUpdate(pusher_name string, branch string, status string, message string) (int64,error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.db.ExecContext(ctx, `INSERT INTO updates (pusher_name, branch, status, message) VALUES (?, ?, ?, ?)`, pusher_name, branch, status, message)
	if err != nil {
		return 0,err
	}

	id,err := result.LastInsertId()

	if(err != nil){
		return 0,err
	}

	return id,nil
}

func (s * service) UpdateStatusAndMessage(id int64, status string,message string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx, `UPDATE updates SET status = ?,message = ? WHERE id = ?`, status,message,id)
	if err != nil {
		fmt.Println("error in update status and message",err)
		return err
	}

	return nil
}


func (s *service) GetUpdates(limit int ,offset int) ([]Update, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	
	rows, err := s.db.QueryContext(ctx, `SELECT * FROM updates ORDER BY id DESC, created_at DESC LIMIT ? OFFSET ?`, limit, offset)

	if err != nil {
		fmt.Println("error in query",err)
		return nil, err
	}

	defer rows.Close()

	var updates []Update
	for rows.Next() {
		var update Update
		err := rows.Scan(&update.ID, &update.PusherName, &update.Branch, &update.Status, &update.Message, &update.CreatedAt, &update.UpdatedAt)
		if err != nil {
			fmt.Println("error",err)
			return nil, err
		}

		updates = append(updates, update)
	}

	return updates, nil
}