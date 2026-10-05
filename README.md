# rssgator

🚧 WIP 🚧


## Usage

Create a `.rssgatorconfig.json` file in your home directory (`$HOME`).

```sh
cd rssgator
go run .
```

## Updating database and models

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
