package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v4"
	"github.com/manniwood/dbmigrator"
)

func main() {
	// postgresql://[user[:password]@][netloc][:port][,...][/dbname][?param1=value1&...]
	// eg
	// DATABASE_URL=postgresql://postgres:postgres@localhost:5432/postgres?application_name=migrator
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Unable to connect to PostgreSQL: %v\n", err)
	}
	defer conn.Close(context.Background())

	dir := os.Getenv("MIGRATIONS_DIR")

	migrator, err := dbmigrator.New(conn, dir)
	if err != nil {
		log.Fatalf("Unable to set up migrator: %v\n", err)
	}
	err = migrator.Migrate(context.Background())
	if err != nil {
		log.Fatalf("Unable to migrate: %v\n", err)
	}
}
