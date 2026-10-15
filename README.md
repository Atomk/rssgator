# rssgator

A simple RSS client.

Built as an exercise for one of [Boot.dev](https://www.boot.dev/)'s Go courses.


## Usage

Create a `.rssgatorconfig.json` file in your home directory (`$HOME`).

```sh
cd rssgator
go run .
```


## Setup

This tool expects you to have Go and PostgreSQL installed on your machine.

### Go

See: https://go.dev/doc/install

### Postgres
Instructions vary by system, for reference on Ubuntu (WSL) I used this:
https://learn.microsoft.com/en-us/windows/wsl/tutorials/wsl-database#install-postgresql

```sh
sudo apt update
sudo apt install postgresql postgresql-contrib
# Assuming installation was successful, start the database server
sudo service postgresql start
```

After setting up postgres you need to create the database used by this program:
```sh
sudo -u postgres psql
# In the psql shell
CREATE DATABASE rssgator;
\c rssgator
ALTER USER postgres PASSWORD 'mypassword';
```
The name of the database can be whatever you want, I used `rssgator`. It's not necessary to set a password for the DB user if you just use this locally but it's recommended. Take note of the database name and password you set here.

### Configuration
When you first run the program, a configuration file named `.rssgatorconfig.json` will be created in your home directory (`$HOME`), the program will print the exact path. Open that file, and update the `db_url` field with a connection string for the database you created: `postgres://postgres:<password>!@localhost:5432/<database>?sslmode=disable` (replace password and database with your own)
The `sslmode` parameter is needed because Go's postgres driver enables SSL by default, but since we use a local database it may result in errors. A plain text connection is fine.

After this, you won't need to touch this file anymore.


## Updating database and models

You need:
- [sqlc](https://sqlc.dev/): `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
- [goose](https://github.com/pressly/goose): `go install github.com/pressly/goose/v3/cmd/goose@latest`

When adding a new migration, you need to:
1. Run `goose` to update the database (after replacing auth in the connection string)
    ```sh
    cd sql/schema
    # the `sslmode` parameter is not required in this connection string
    goose postgres 'postgres://postgres:<password>@localhost:5432/rssgator' up
    # 2026/10/06 00:59:47 OK   002_feeds.sql (45.15ms)
    # 2026/10/06 00:59:47 goose: successfully migrated database to version: 2
    ```
2. Run `sql` to create/update the corresponding Go functions and models:
    ```sh
    cd -
    sqlc generate
    ```
