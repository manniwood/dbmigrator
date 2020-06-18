package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v4"
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

	var numeral int
	err = conn.QueryRow(context.Background(), "select 1").Scan(&numeral)
	if err != nil {
		log.Fatalf("QueryRow failed: %v\n", err)
	}

	fmt.Println(numeral)
}
