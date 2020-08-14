// Package dbmigrator migrates PostgreSQL databases
package dbmigrator

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

// advLockClassID is a magic number chosen to identify a PostgreSQL avisory lock
// that is used to prevent more than one copy of dbmigrator running at the same time.
const advLockClassID pgtype.OID = 12345
// advLockObjectID is a magic number chosen to identify a PostgreSQL avisory lock
// that is used to prevent more than one copy of dbmigrator running at the same time.
const advLockObjectID pgtype.OID = 12345

// NOTE on error handling: we follow the advice at https://blog.golang.org/go1.13-errors:
// The pgx errors we will be dealing with are internal details.
// To avoid exposing them to the caller, we repackage them as new
// errors with the same text. We use the %v formatting verb, since
// %w would permit the caller to unwrap the original pgx errors.
// We don't want to support pgx errors as part of our API.

// Migrator migrates a database from one schema to another.
type Migrator struct {
	// conn is a PostgreSQL connection.
	conn *pgx.Conn
	// dir is the path to the directory where the SQL migration scripts live.
	dir string
}

// New returns a new Migrator. It needs a connection to
// postgres (using a pgx connection) and a directory to
// look for migrations in.
func New(conn *pgx.Conn, dir string) (*Migrator, error) {
	ctx := context.Background()

	// Try to acquire advisory lock. If false, another copy of dbmigrator
	// is running.
	ok, err := tryAdvisoryLock(ctx, conn)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("Another copy of dbmigrator is already running.")
	}

	// Always ensure the migrations table exists first.
	err = createMigrationsTable(ctx, conn)
	if err != nil {
		return nil, err
	}

	return &Migrator{
		conn: conn,
		dir:  dir,
	}, nil
}

// tryAdvisoryLock tries to grab an exclusive advisory lock; if this fails,
// it means another copy of dbmigrator is running and we should stay out of
// its way.
func tryAdvisoryLock(ctx context.Context, conn *pgx.Conn) (bool, error) {
	var ok bool
	err := conn.QueryRow(ctx, "select pg_try_advisory_lock(666, 999)").Scan(&ok)
	if err != nil {
		return false, fmt.Errorf("Problem trying to acquire advisory lock: %v", err)
	}
	return ok, nil
}

// createMigrationsTable idempotently creates the migrations table
// used by dbmigrator. The migration table will not exist on the very first
// migration. The SQL is therefore written to make creating this table
// idempotent.
func createMigrationsTable(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, `
create table if not exists migrations (
  migration text constraint migrations_pk primary key not null,
	applied_on timestamp without time zone not null default now())
`)
	if err != nil {
		return fmt.Errorf("Problem creating migrations table: %v", err)
	}
	return nil
}

// Migrate migrates the db to the most recent migration in
// the directory the migrator was configured to look in.
// Migrate is a bit chatty when it works; however, it chats
// to io.Writer w, which you can make whatever you want, including
// a kind of writer that just throws away messages.
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
	// which are above the current migration. Also, skip
	// any file that does not end with `*.sql`.
	migrationsDone := 0
	for _, f := range files {

		// Skp any file that does not end in ".sql"
		if !strings.HasSuffix(f.Name(), ".sql") {
			continue
		}

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

// doMigration does a single migration listed in fileName.
// It executes the entire file all at once, taking adavantage
// of the fact that PostgreSQL can run more than one (semi-colon-terminated)
// SQL query with a single call.
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
