// Package dbmigrator migrates PostgreSQL databases
package dbmigrator
import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
)
func Migrate(conn *pgx.Conn) error {
	var numeral int
	err := conn.QueryRow(context.Background(), "select 1").Scan(&numeral)
	if err != nil {
		return err
	}

	fmt.Println(numeral)
	return nil
}

/* func Migrate:
1. Find dir with migrate scripts in them.
2. Sort *_up.sql files alphabetically.
3. begin; select max(id) from migrations; rollback;
4. begin;
5. for each sql file greater than max(id)
5.1  apply file
5.2. if error, rollback;, report to user, and os.exit(1)
5.3  insert into migrations (id, apply_time) values (file.id, now())
5.4. if error, rollback;, report to user, and os.exit(1)
6. commit;, report success, and os.exit(0)
*/

/*
apply file: do we just apply the whole damned file, or do we
apply it one statement at a time? What about creation of stored
procedures which has embedded semicolons? Yeah, looks like just
applying the whole damned file is a good way to go, because
Pg allows doing so, and it's way easier to code.
Make a note to the user that "commit;" and "rollback;" are not
allowed in the sql file.
*/
