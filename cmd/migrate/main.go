package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/pressly/goose"
	_ "github.com/lib/pq"
)

func main() {
	migrationsDir := flag.String("dir", "migrations", "Directory with migration files")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		log.Fatal("Usage: migrate <command> [arguments]\nCommands: up, down, status, redo")
	}

	command := args[0]
	dbString := os.Getenv("DATABASE_URL")
	if dbString == "" {
		// Default connection string matching docker-compose.yml
		dbString = "postgres://postgres:postgres@localhost:5488/app?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbString)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	switch command {
	case "up":
		if err := goose.Up(db, *migrationsDir); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
		fmt.Println("Migrations applied successfully")
	case "down":
		if err := goose.Down(db, *migrationsDir); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
		fmt.Println("Migration rolled back successfully")
	case "status":
		if err := goose.Status(db, *migrationsDir); err != nil {
			log.Fatalf("Migration status failed: %v", err)
		}
	case "redo":
		if err := goose.Redo(db, *migrationsDir); err != nil {
			log.Fatalf("Migration redo failed: %v", err)
		}
		fmt.Println("Migration redone successfully")
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}
