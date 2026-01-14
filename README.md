# Polling App

A real-time polling application built with Go, PostgreSQL, and Ent ORM.

## Features

- Get/Delete/List Polls
- Vote on a Poll (via UI and API)
- Real-time vote counting
- Database migrations with Atlas

## Tech Stack

- **Backend**: Go 1.21+
- **Database**: PostgreSQL 16
- **ORM**: Ent
- **Migrations**: Atlas
- **Containerization**: Docker

## Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Make

### Setup

1. **Clone and setup the database:**

```bash
# Start PostgreSQL
make db-up

# Run migrations
make migrate-apply
```

2. **Run the application:**

```bash
make run
```

3. **Access the app:**
   - API: http://localhost:8080
   - Health check: http://localhost:8080/health

## Available Commands

### Development
```bash
make help              # Show all available commands
make build             # Build the application
make run               # Run the application
make test              # Run tests
make lint              # Run linters
make fmt               # Format code
```

### Database
```bash
make db-up             # Start PostgreSQL container
make db-down           # Stop PostgreSQL container
make db-shell          # Connect to database shell
make seed              # Seed database with demo data
```

### Migrations

```bash
# Check migration status
make migrate-status

# Apply pending migrations
make migrate-apply

# Create new migration (requires Docker)
make migrate-new name=add_feature

# Validate migration files
make migrate-validate

# Recalculate hash (after manually removing migrations)
make migrate-hash

# Run full CI/CD migration pipeline
make migrate-ci
```

## Database Migrations

This project uses [Atlas](https://atlasgo.io/) for database migrations with [Ent](https://entgo.io/) schema definitions.

### For Local Development

```bash
# 1. Start database
make db-up

# 2. Apply migrations
make migrate-apply

# 3. Check status
make migrate-status
```

### For CI/CD

Use the single `migrate-ci` command in your pipeline:

```yaml
# GitHub Actions / GitLab CI
- name: Run Migrations
  env:
    DB_HOST: ${{ secrets.DB_HOST }}
    DB_PASSWORD: ${{ secrets.DB_PASSWORD }}
  run: make migrate-ci
```

This command:
- ✅ Auto-installs Atlas if needed
- ✅ Validates migrations
- ✅ Shows status
- ✅ Applies pending migrations

### Configuration

Set these environment variables for different environments:

```bash
DB_HOST=localhost          # Database host
DB_PORT=5432              # Database port
DB_USER=polling           # Database user
DB_PASSWORD=polling       # Database password
DB_NAME=polling_app       # Database name
```

### Creating New Migrations

When you modify Ent schemas in `internal/ent/schema/`, create a new migration:

```bash
make migrate-new name=add_user_role
```

**Note**: Requires Docker. If you have permission issues:

```bash
# Use a separate dev database
export ATLAS_DEV_URL="postgres://user:pass@localhost:5432/dev_db?sslmode=disable"
make migrate-new name=add_user_role
```

### Removing a Migration

#### If Migration Is NOT Yet Applied (Pending)

```bash
# 1. Check status to confirm it's pending
make migrate-status

# 2. Delete the migration file
rm internal/migrate/migrations/YYYYMMDDHHMMSS_migration_name.sql

# 3. Recalculate the hash
make migrate-hash
```

#### If Migration Is Already Applied

You have two options:

**Option 1: Create a Reverse Migration (Recommended for Production)**

```bash
# Create a new migration that undoes the changes
make migrate-new name=remove_feature_xyz

# Edit the generated migration to reverse the changes
# Then apply it
make migrate-apply
```

**Option 2: Use Declarative Approach**

```bash
# Modify your Ent schema to remove the changes
# Then create a new migration
make migrate-new name=cleanup_xyz

# This will detect the differences and create the appropriate migration
```

**⚠️ Option 3: Reset Database (DESTRUCTIVE - Development Only)**

```bash
# This will delete ALL data and reapply migrations
make migrate-reset
```

### Troubleshooting

#### Docker permission denied

```bash
# Option 1: Fix docker permissions
sudo usermod -aG docker $USER
newgrp docker

# Option 2: Use separate dev database
export ATLAS_DEV_URL="postgres://user:pass@localhost:5432/dev_db?sslmode=disable"
make migrate-new name=my_migration
```

#### Atlas not found

The Makefile auto-installs Atlas to `~/bin`. Or install manually:

```bash
make install-atlas
```

## Project Structure

```
.
├── cmd/
│   ├── server/          # Main application entry point
│   └── seed/            # Database seeding utility
├── internal/
│   ├── ent/             # Ent ORM generated code (auto-generated)
│   │   ├── schema/      # Entity definitions (user-defined, edit these)
│   │   ├── *.go         # Generated code (do not edit)
│   │   └── */           # Generated packages (do not edit)
│   └── migrate/
│       └── migrations/  # Database migration files (version controlled)
├── Makefile             # Build and development commands
└── docker-compose.yml   # Local development services
```

### Ent Code Generation

**Schema files** (edit these):
- Location: `internal/ent/schema/*.go`
- Define your entities (User, Poll, Vote, etc.)
- Version controlled

**Generated files** (do not edit):
- Location: `internal/ent/*.go` and `internal/ent/*/`
- Auto-generated by running `make ent-gen`
- Can be version controlled or gitignored (your choice)

To regenerate Ent code after modifying schemas:
```bash
make ent-gen
```

## Development Workflow

1. **Modify Ent schemas** in `internal/ent/schema/*.go` (User, Poll, Vote, etc.)
2. **Generate Ent code**: `make ent-gen` (generates code in `internal/ent/`)
3. **Create migration**: `make migrate-new name=description`
4. **Review generated SQL** in `internal/migrate/migrations/`
5. **Apply locally**: `make migrate-apply`
6. **Test your changes**
7. **Commit** schema files, generated code (optional), and migration files

### Working with Ent

```bash
# After modifying internal/ent/schema/*.go files
make ent-gen              # Regenerate Ent code

# Then create database migration
make migrate-new name=add_user_field

# Apply the migration
make migrate-apply
```

## CI/CD Integration

Example GitHub Actions workflow:

```yaml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  migrate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      
      - name: Run Migrations
        env:
          DB_HOST: ${{ secrets.DB_HOST }}
          DB_PORT: ${{ secrets.DB_PORT }}
          DB_USER: ${{ secrets.DB_USER }}
          DB_PASSWORD: ${{ secrets.DB_PASSWORD }}
          DB_NAME: ${{ secrets.DB_NAME }}
        run: make migrate-ci
  
  deploy:
    needs: migrate
    runs-on: ubuntu-latest
    steps:
      - name: Deploy application
        run: # your deployment steps
```

## License

MIT
