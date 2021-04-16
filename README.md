[![](https://godoc.org/github.com/manniwood/dbmigrator?status.svg)](https://godoc.org/github.com/manniwood/dbmigrator)
[![Build Status](https://travis-ci.com/manniwood/dbmigrator.svg)](https://travis-ci.com/manniwood/dbmigrator)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# DBMigrator: A Minimalist Database Migrator

## Project Status: Experimental

Just playing with some ideas here; figured I'd share. I wouldn't reccomend
using this in production, but I'd be pleased if anybody stole some ideas
from this project: it's quite permissively licenced.

## Summary

DBMigrator is an experiment to see just how minimal a database migrator can be.

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

Your migration scripts must all end in '.sql', or they will be ignored.

Your migration scripts are assumed to be applied in alphabetical order,
so choose a naming scheme for your '.sql' files that keeps them alphabetical
by the order you wish the migrations to happen in.

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

## Testing

You need a postgresql instance at localhost:5432 with the default
database (postgres) and user (postgres) containing no tables.

Running the test suite will leave behind tables in the postgres
database, which will have to be deleted manually.

## Future plans

I have very few other plans for this! I wanted to prove to myself just how minimalistic
a database migrator could be and still be useful.

I hope the licence is permissive enough that if anybody wants to use this code as
inspiration for a similar project (maybe a MySQL-only version, for instance), anybody
can feel free to go ahead and do so.


