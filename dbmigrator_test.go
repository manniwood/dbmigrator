package dbmigrator
import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v4"
)

const TestDBURL string = "postgresql://postgres:postgres@localhost:5432/postgres?application_name=migrator"

// We need to clean up after ourselves so that our test state is OK
func dropTables(t *testing.T, tables ...string) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, TestDBURL)
	if err != nil {
		t.Errorf("CLEANUP: Problem trying to connect to Pg so that we can drop tables: %v", err)
	}

	for _, table := range tables {
		_, err := conn.Exec(ctx, fmt.Sprintf(`drop table if exists %s`, table))
		if err != nil {
			t.Errorf("CLEANUP: Problem trying to drop table %s: %v", table, err)
		}
	}
}

func TestMigrate(t *testing.T) {
	defer dropTables(t, "a", "b", "c", "migrations")

	// The following migrations all build on each other
	// and have to be executed in the order shown; do not
	// run in parallel.

	tests := []struct {
		name           string
		dir            string
		wantLog        string
		wantTables     []string
		wantMigrations []string
	}{
		{
			name:           "Empty migrations dir.",
			dir:            "test_migrations/empty",
			wantLog:        "No migrations yet",
			wantTables:     []string{},
			wantMigrations: []string{},
		},
		{
			name:           "Migrations dir with a.sql.",
			dir:            "test_migrations/1",
			wantLog:        "Did 1 migrations",
			wantTables:     []string{"a"},
			wantMigrations: []string{"a.sql"},
		},
		{
			name:           "Migrations dir with a.sql, b.sql, c.sql; a.sql already applied.",
			dir:            "test_migrations/2",
			wantLog:        "Did 2 migrations",
			wantTables:     []string{"b", "c"},
			wantMigrations: []string{"b.sql", "c.sql"},
		},
	}

	for _, tt := range tests {
		conn, err := pgx.Connect(context.Background(), TestDBURL)
		if err != nil {
			t.Errorf("Unable to connect to PostgreSQL: %v", err)
		}
		migrator, err := New(conn, tt.dir)
		if err != nil {
			t.Errorf("Unable to set up migrator: %v", err)
		}
		// Listen to logs from the migrator.
		buf := bytes.Buffer{}
		err = migrator.Migrate(context.Background(), &buf)
		if err != nil {
			t.Errorf("Unable to migrate: %v", err)
		}
		if !strings.Contains(buf.String(), tt.wantLog) {
			t.Errorf("Logs expected >>>%s<<<; got >>>%s<<< instead.", tt.wantLog, buf.String())
		}

		// Be sure the migrations made it to the migrations table.
		for _, migration := range tt.wantMigrations {
			var ct int
			err = conn.QueryRow(context.Background(), "select count(*) from migrations where migration = $1", migration).Scan(&ct)
			if err != nil {
				t.Errorf("Could not get migrations row count for %s: %v", migration, err)
			}
			if ct != 1 {
				t.Errorf("%s was not found in the migrations table.", migration)
			}
		}

		// Be sure the tables actually exist in PostgreSQL.
		for _, table := range tt.wantTables {
			var ct int
			err = conn.QueryRow(context.Background(), "select count(*) from pg_class where relname = $1", table).Scan(&ct)
			if err != nil {
				t.Errorf("Problem looking for table %s in pg_class: %v", table, err)
			}
			if ct != 1 {
				t.Errorf("Could not find table %s after migration.", table)
			}
		}
		err = conn.Close(context.Background())
		if err != nil {
			t.Errorf("Problem closing connection to pg; possibility of advisory lock not being closed: %v", err)
		}
	}

}

func TestAdvisoryLock(t *testing.T) {
	defer dropTables(t, "migrations")
	conn, err := pgx.Connect(context.Background(), TestDBURL)
	if err != nil {
		t.Errorf("Unable to connect to PostgreSQL: %v", err)
	}
	defer conn.Close(context.Background())

	_, err = New(conn, "test_migrations/empty")
	if err != nil {
		t.Errorf("Unable to set up migrator: %v", err)
	}

	conn2, err := pgx.Connect(context.Background(), TestDBURL)
	if err != nil {
		t.Errorf("Unable to conn2ect to PostgreSQL: %v", err)
	}
	defer conn2.Close(context.Background())
	_, err = New(conn2, "test_migrations/empty")
	if err == nil {
		t.Errorf("Second migrator was supposed to throw an error, but it did not!")
	}
}
