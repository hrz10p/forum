run:
	@go run ./cmd/

migrate:
	@sqlite3 forum.sqlite < ./pkg/migrations/tables.sql
	@echo "Migrated tables"

