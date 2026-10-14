# rssgator

🚧 WIP 🚧


## Usage

Create a `.rssgatorconfig.json` file in your home directory (`$HOME`).

```sh
cd rssgator
go run .
```


## Setup

This tool expects you to have PostgreSQL installed on your machine.
Instructions vary by system, for reference on Ubuntu (WSL) I used this:
https://learn.microsoft.com/en-us/windows/wsl/tutorials/wsl-database#install-postgresql

```sh
sudo apt update
sudo apt install postgresql postgresql-contrib
```

After setting up postgres you need to create the database used by this program:
```sh
sudo -u postgres psql
# In the psql shell
CREATE DATABASE rssgator;
\c rssgator
ALTER USER postgres PASSWORD 'mypassword';
```
The name of the database can be whatever you want, I used `rssgator`. It's not necessary to set a password for the DB user if you just use this locally but it's recommended. Note that you'll need the database name and password set here to put them in the configuration file and the connection string used with the `goose` command to run migrations.


## Updating database and models

You need:
- [sqlc](https://sqlc.dev/): `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
- [goose](https://github.com/pressly/goose): `go install github.com/pressly/goose/v3/cmd/goose@latest`

When adding a new migration, you need to:
1. Run `goose` to update the database (after replacing auth in the connection string)
    ```sh
    cd sql/schema
    goose postgres 'postgres://myuser:mypassword@localhost:5432/rssgator' up
    # 2026/10/06 00:59:47 OK   002_feeds.sql (45.15ms)
    # 2026/10/06 00:59:47 goose: successfully migrated database to version: 2
    ```
2. Run `sql` to create/update the corresponding Go functions and models:
    ```sh
    cd -
    sqlc generate
    ```
