.PHONY: run migrate-up migrate-down test swagger docker-up

run:
	go run cmd/server/main.go

migrate-up:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/subscriptions?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/subscriptions?sslmode=disable" down

test:
	go test -v ./... -cover

swagger:
	swag init -g cmd/server/main.go -o docs

docker-up:
	docker-compose up --build