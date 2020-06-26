package dbmigrator
import (
	"bytes"
	"context"
	"testing"

	"github.com/jackc/pgx/v4"
)

func TestMigrate(t *testing.T) {
	dbURL := "postgresql://postgres:postgres@localhost:5432/postgres?application_name=migrator"
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		t.Errorf("Unable to connect to PostgreSQL: %v", err)
	}
	defer conn.Close(context.Background())
	if err != nil {
		t.Errorf("Error getting item: %v", err)
	}

	dir := "test_migrations/empty"

	migrator, err := New(conn, dir)
	if err != nil {
		t.Errorf("Unable to set up migrator: %v", err)
	}

	// Listen to logs from the migrator.
	var buf bytes.Buffer
	err = migrator.Migrate(context.Background(), &buf)
	if err != nil {
		t.Errorf("Unable to migrate: %v", err)
	}
	if buf.String() != "No migrations yet.\nDid 0 migrations.\n" {
		t.Errorf("Expected no migrations message; got >>>%s<<< instead.", buf.String())
	}

}
