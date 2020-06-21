// Package dbmigrator migrates PostgreSQL databases
package dbmigrator

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/jackc/pgx/v4"
)

// NOTE on error handling: we follow the advice at https://blog.golang.org/go1.13-errors:
// The pgx errors we will be dealing with are internal details.
// To avoid exposing them to the caller, we repackage them as new
// errors with the same text. We use the %v formatting verb, since
// %w would permit the caller to unwrap the original pgx errors.
// We don't want to support pgx errors as part of our API.

// createMigrationsTable is the SQL to ensure the
// migrations table exists (because it won't
// on the very first migration). The SQL
// is therefore written to make creating this
// table idempotent.
const createMigrationsTable string = `
create table if not exists migrations (
  migration text constraint migrations_pk primary key not null,
	applied_on timestamp without time zone not null default now())
`
// Migrator migrates a database from one schema to another.
type Migrator struct {
	// conn is a PostgreSQL connection.
	conn *pgx.Conn
	// dir is the path to the directory where the SQL migration scripts live.
	dir string
}

func New(conn *pgx.Conn, dir string) (*Migrator, error) {
	// Always ensure the migrations table exists first.
	ctx := context.Background()
	_, err := conn.Exec(ctx, createMigrationsTable)
	if err != nil {
		return nil, fmt.Errorf("Problem creating migrations table: %v", err)
	}
	return &Migrator{
		conn: conn,
		dir:  dir,
	}, nil
}

func (m *Migrator) Migrate(ctx context.Context, w io.Writer) error {
	// Get the current status of the database by getting the maximum migration
	// from the migrations table.
	var current string
	err := m.conn.QueryRow(ctx, `
    select coalesce(max(migration), '') as current
      from migrations
	`).Scan(&current)
	if err != nil {
		return fmt.Errorf("Problem getting current status: %v", err)
	}
	if current == "" {
		fmt.Fprintf(w, "No migrations yet.\n")
	} else {
		fmt.Fprintf(w, "Current database state: %s\n", current)
	}

	// Get the list of database migrations in alphabetical order.
	files, err := ioutil.ReadDir(m.dir)
	if err != nil {
		return fmt.Errorf("Problem reading directory %s: %v", m.dir, err)
	}
	// Usefully, ioutilReadDir() returns a sorted list!
	// Go through the list and only hold on to those
	// which are above the current migration.
	migrationsDone := 0
	for _, f := range files {
		if strings.Compare(f.Name(), current) == 1 {
			// This migration has not been done yet.
			fmt.Fprintf(w, "Migrating %s\n", f.Name())
			err = m.doMigration(ctx, f.Name())
			if err != nil {
				return fmt.Errorf("Problem migrating %s: %v", f.Name(), err)
			}
			// Update migrations table with f.Name()
			_, err = m.conn.Exec(ctx, `insert into migrations
					(migration) values ($1)`, f.Name())
			if err != nil {
				return fmt.Errorf("Problem updating migrations table with %s: %v", f.Name(), err)
			}
			migrationsDone++
		}
	}
	fmt.Fprintf(w, "Did %d migrations.\n", migrationsDone)

	return nil
}

func (m *Migrator) doMigration(ctx context.Context, fileName string) error {
	fileBytes, err := ioutil.ReadFile(m.dir + "/" + fileName)
	if err != nil {
		return fmt.Errorf("Problem reading file %s: %v", fileName, err)
	}
	_, err = m.conn.Exec(ctx, string(fileBytes))
	if err != nil {
		return fmt.Errorf("Problem running SQL from %s: %v", fileName, err)
	}
	return nil
}
