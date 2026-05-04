# Subscription Service

REST API for managing user subscriptions (test task for Effective Mobile).

## Features
- CRUDL operations for subscriptions
- Total cost calculation for a period with filters (user_id, service_name)
- PostgreSQL with migrations
- Configuration via YAML + environment variables
- Swagger documentation
- Docker Compose for easy setup
- Structured logging, graceful shutdown, request ID middleware

## Tech stack
- Go 1.24, Echo, sqlx, PostgreSQL
- Migrations: golang-migrate
- Tests: testcontainers, testify
- Swagger: swaggo

## Quick start
```bash
# Clone
git clone https://github.com/yourusername/subscription-service.git
cd subscription-service

# Run with Docker Compose
make docker-up

# Swagger UI: http://localhost:8080/swagger/index.html