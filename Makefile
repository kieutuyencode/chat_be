MIGRATION_NAME ?= migration_name
MIGRATION_DIR = "file://database/ent/migrate/migrations"

SCHEMA_DIR = "ent://database/ent/schema"

DEV_URL = "docker://postgres/15/test?search_path=public"

DB_URL = "postgres://root:secret@localhost:5434/chat_be?search_path=public&sslmode=disable"


migrate_generate:
	atlas migrate diff $(MIGRATION_NAME) \
	--dir $(MIGRATION_DIR) \
	--to $(SCHEMA_DIR) \
	--dev-url $(DEV_URL)

migrate_apply:
	atlas migrate apply \
	--dir $(MIGRATION_DIR) \
	--url $(DB_URL)

migrate_status:
	atlas migrate status \
	--dir $(MIGRATION_DIR) \
	--url $(DB_URL)

schema_inspect:
	atlas schema inspect \
	-u $(SCHEMA_DIR) \
	--dev-url $(DEV_URL) \
	-w

schema_generate:
	go generate ./database/ent

server:
	go run cmd/server/server.go