# Polling App

A polling application built with Go and PostgreSQL.

## Features

- Register users
- Create/Get/Delete/List Polls
- Vote on a Poll

## Out of Scope

- Authentication/Authorization

## Tech Stack

- Go standard library (net/http)
- Ent ORM (with auto-migrations)
- Testcontainers for integration tests

## Dependencies

- Go 1.21+
- PostgreSQL 16
- Docker (for local setup and integration tests)

## Local Setup

```bash
# Start PostgreSQL
make db-up

# Run migrations
make migrate-apply

# Run the application
make run
```

## Development

```bash
# Install pre-commit hooks (runs fmt + lint on commit)
make install-hooks

# Format code
make fmt

# Run linters
make lint

# Run unit tests
make test

# Run integration tests (requires Docker)
make test-integration

# For all available commands
make help
```

## API Usage

### Health Check

```bash
curl http://localhost:8080/health
```

### Register Users

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "email": "alice@example.com"}'
```

### Create Polls

```bash
curl -X POST http://localhost:8080/polls \
  -H "Content-Type: application/json" \
  -d '{"owner_id": 1, "title": "Best Language?", "options": ["Go", "Rust", "Python"]}'
```

### List & Get Polls

```bash
curl http://localhost:8080/polls
curl http://localhost:8080/polls/1
```

### Vote on a Poll

```bash
curl -X POST http://localhost:8080/polls/1/vote \
  -H "Content-Type: application/json" \
  -d '{"option_id": 1, "user_id": 1}'
```

### Delete a Poll

```bash
curl -X DELETE http://localhost:8080/polls/1
```
