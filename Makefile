include .env
export

DB_URL=postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/$(POSTGRES_DB)?sslmode=disable

.PHONY: migrateup migratedown migratestatus migrateforce sqlc

# Show current migration version
migratestatus:
	@echo "Checking migration status..."
	migrate -path db/migrate -database "$(DB_URL)" version

# Run migrations up
migrateup:
	@echo "Running migrations up..."
	migrate -path db/migrate -database "$(DB_URL)" up
	@echo "Current migration version:"
	@make migratestatus

# Run migrations down
migratedown:
	@echo "Running migrations down..."
	migrate -path db/migrate -database "$(DB_URL)" down
	@echo "Current migration version:"
	@make migratestatus

# Force set migration version (use with caution)
migrateforce:
	@echo "Forcing migration version $(version)..."
	migrate -path db/migrate -database "$(DB_URL)" force $(version)
	@echo "Current migration version:"
	@make migratestatus

sqlc:
	sqlc generate