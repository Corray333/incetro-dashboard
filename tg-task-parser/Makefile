include .env

.SILENT:
build:
	cd cmd && go build main.go
run: build
	cd cmd && ./main
	
goose-up:
	cd migrations && goose postgres "user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) host=localhost port=5432 dbname=$(POSTGRES_DB_NAME) sslmode=disable" up
goose-down:
	cd migrations && goose postgres "user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) host=localhost port=5432 dbname=$(POSTGRES_DB_NAME) sslmode=disable" down
goose-down-all:
	cd migrations && goose postgres "user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) host=localhost port=5432 dbname=$(POSTGRES_DB_NAME) sslmode=disable" down-to 0