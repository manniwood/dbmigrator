package dbmigrator
import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v4"
)

const TestDBURL string = "postgresql://postgres:postgres@localhost:5432/postgres?application_name=migrator"

func TestMigrate(t *testing.T) {
	conn, err := pgx.Connect(context.Background(), TestDBURL)
	if err != nil {
		t.Errorf("Unable to connect to PostgreSQL: %v", err)
	}
	defer conn.Close(context.Background())
	if err != nil {
		t.Errorf("Error getting item: %v", err)
	}

	// The following migrations all build on each other.

	// ---- test_migrations/empty
	migrator, err := New(conn, "test_migrations/empty")
	if err != nil {
		t.Errorf("Unable to set up migrator: %v", err)
	}
	// Listen to logs from the migrator.
	var buf bytes.Buffer
	err = migrator.Migrate(context.Background(), &buf)
	if err != nil {
		t.Errorf("Unable to migrate: %v", err)
	}
	if !strings.Contains(buf.String(), "No migrations yet") {
		t.Errorf("Expected no migrations message; got >>>%s<<< instead.", buf.String())
	}

	// ---- test_migrations/1

	migrator, err = New(conn, "test_migrations/1")
	if err != nil {
		t.Errorf("Unable to set up migrator: %v", err)
	}
	// Listen to logs from the migrator.
	buf = bytes.Buffer{}
	err = migrator.Migrate(context.Background(), &buf)
	if err != nil {
		t.Errorf("Unable to migrate: %v", err)
	}

	var ct int
	err = conn.QueryRow(context.Background(), "select count(*) from migrations").Scan(&ct)
	if err != nil {
		t.Errorf("Could not get migrations row count: %v", err)
	}
	if ct != 1 {
		t.Errorf("Expected 1 migration in migrations table, found %v instead.", ct)
	}

	err = conn.QueryRow(context.Background(), "select count(*) from migrations where migration = 'a.sql'").Scan(&ct)
	if err != nil {
		t.Errorf("Could not get migrations row count for a.sql: %v", err)
	}
	if ct != 1 {
		t.Errorf("a.sql was not found in the migrations table.")
	}

	err = conn.QueryRow(context.Background(), "select count(*) from pg_class where relname = 'a'").Scan(&ct)
	if err != nil {
		t.Errorf("Problem looking for table a in pg_class: %v", err)
	}
	if ct != 1 {
		t.Errorf("Could not find table a after migrating a.sql.")
	}

	// ---- test_migrations/2

	migrator, err = New(conn, "test_migrations/2")
	if err != nil {
		t.Errorf("Unable to set up migrator: %v", err)
	}
	// Listen to logs from the migrator.
	buf = bytes.Buffer{}
	err = migrator.Migrate(context.Background(), &buf)
	if err != nil {
		t.Errorf("Unable to migrate: %v", err)
	}

	ct = 0
	err = conn.QueryRow(context.Background(), "select count(*) from migrations").Scan(&ct)
	if err != nil {
		t.Errorf("Could not get migrations row count: %v", err)
	}
	if ct != 3 {
		t.Errorf("Expected 3 migration in migrations table, found %v instead.", ct)
	}

	err = conn.QueryRow(context.Background(), "select count(*) from migrations where migration in ('b.sql', 'c.sql')").Scan(&ct)
	if err != nil {
		t.Errorf("Could not get migrations row count for b.sql and c.sql: %v", err)
	}
	if ct != 2 {
		t.Errorf("b.sql and/or c.sql were not found in the migrations table.")
	}

	err = conn.QueryRow(context.Background(), "select count(*) from pg_class where relname in ('b', 'c')").Scan(&ct)
	if err != nil {
		t.Errorf("Problem looking for tables b and c in pg_class: %v", err)
	}
	if ct != 2 {
		t.Errorf("Could not find tables b and/or c after migrating b.sql and c.sql.")
	}
}
