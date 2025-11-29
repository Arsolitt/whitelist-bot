```shell
docker compose up -d
```

```shell
goose sqlite3 data/whitelist.db -dir migrations/sqlite up
```

```shell
goose postgres "postgres://app:app@127.0.0.1:5432/app" -dir migrations/postgres up
```

```shell
goose sqlite3 data/whitelist.db -dir migrations/sqlite create create_users_table sql
```

```shell
goose postgres "user=postgres dbname=postgres sslmode=disable" -dir migrations/postgres create create_users_table sql
```

```shell
sqlc generate
```

```shell
go test ./...
```
