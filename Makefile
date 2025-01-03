DB_URL=postgresql://bank_app_user:BankApp123@localhost:5432/bank_app_database?sslmode=disable

.PHONY: migrateup migratedown migratestatus migrateforce sqlc test

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

test:
	go test -v -cover ./...