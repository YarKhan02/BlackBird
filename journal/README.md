# Database Migration

```sh
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Then run the following command to apply the migrations:

```sh
./bin/db-migrate.sh
```