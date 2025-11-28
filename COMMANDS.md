```shell
docker compose up -d
```

```shell
goose sqlite3 data/whitelist.db -dir migrations/sqlite up
```

```shell
goose sqlite3 data/whitelist.db -dir migrations/sqlite create create_users_table sql
```

```shell
sqlc generate
```

```shell
go test ./...
```
