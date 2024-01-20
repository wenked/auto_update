package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type Service interface {
	Health() map[string]string
}

type service struct {
	db *sql.DB
}

var (
	dburl = os.Getenv("DB_URL")
)

func New() Service {
	db, err := sql.Open("libsql", dburl)
	if err != nil {
		// This will not be a connection error, but a DSN parse error or
		// another initialization error.
		log.Fatal(err)
	}

	fmt.Println("dburl", dburl)


	
	_,migration_err := db.Exec(`CREATE TABLE IF NOT EXISTS "updates" (
		ID INTEGER PRIMARY KEY AUTOINCREMENT, 
		"pusher_name" TEXT, 
		"branch" TEXT, 
		"status" TEXT, 
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

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := s.db.PingContext(ctx)
	if err != nil {
		log.Fatalf(fmt.Sprintf("db down: %v", err))
	}

	return map[string]string{
		"message": "It's healthy",
	}
}
