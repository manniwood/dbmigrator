[![](https://godoc.org/github.com/manniwood/dbmigrator?status.svg)](https://godoc.org/github.com/manniwood/dbmigrator)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# DBMigrator (Work in Progress)

DBMigrator is an experiment to see just how simple a database migrator can be.

I was tempted to just write this in Bash, calling PostgreSQL's `psql` client
when necessary. But I decided to do it in Go.

## Limitations

DBMigrator only works with PostgreSQL.

DBMigrator only does forward migrations. (If you need to roll back a migration,
you need to make a migration that undoes the migration you want to undo.)

## Setup

```
git clone git@github.com:manniwood/dbmigrator.git
cd dbmigrator/cmd/migrate
go build
```

## Usage

Your migration scripts will be written in plain-old SQL (PostgreSQL's flavor).

Your migration scripts must all live together in the same directory.

Only migration scripts are allowed to live in your migrations directory.

Migrations only work forward. If you need to undo a previous migration,
create a new migration that undoes the migration you need to undo.

Example run on an empty `postgres` database, using the sample migrations
in `../example`. (NOTE that the database and user
named `postgres` come with default PostgreSQL installations, following a
convention that this is a db for tools to connect to.)

```
$ DATABASE_URL=postgresql://postgres:postgres@localhost:5432/postgres?application_name=migrator \
  MIGRATIONS_DIR=../../example \
  ./migrate
No migrations yet.
Migrating a.sql
Migrating b.sql
Migrating c.sql
Did 3 migrations.
```
What happens if we run again on the same database:

```
$ DATABASE_URL=postgresql://postgres:postgres@localhost:5432/postgres?application_name=migrator \
  MIGRATIONS_DIR=../../example \
  ./migrate
Current database state: c.sql
Did 0 migrations.
```
What happens if we add two new migrations, `d.sql` and `e.sql` to the migrations directory:

```
$ DATABASE_URL=postgresql://postgres:postgres@localhost:5432/postgres?application_name=migrator \
  MIGRATIONS_DIR=../../example \
  ./migrate
Current database state: c.sql
Migrating d.sql
Migrating e.sql
Did 2 migrations.
```
What happens if we run again on the same database:

```
$ DATABASE_URL=postgresql://postgres:postgres@localhost:5432/postgres?application_name=migrator \
  MIGRATIONS_DIR=../../example \
  ./migrate
Current database state: e.sql
Did 0 migrations.
```
## Future plans

I plan to write an integration test suite for this.

I have very few other plans for this! I wanted to prove to myself just how much
a database migrator could be under-engineered and still be useful.

I hope the licence is permissive enough that if anybody wants to use this code as
inspiration for a similar project (maybe a MySQL-only version, for instance), anybody
can feel free to go ahead and do so.

