# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is `anime-sync`, a Go application that syncs anime data from various sources (primarily Pulsar/Kafka) to a MySQL database. It's designed as a POC since pulsar-mariadb sync does not work directly.

## Architecture

The application follows a layered architecture:

- **CLI Commands**: Located in `internal/commands/` using Cobra CLI framework
- **Database Layer**: GORM-based repositories in `internal/db/repositories/`
- **Event Processing**: Generic processor pattern in `internal/services/processor/`  
- **Message Consumption**: Pulsar and Kafka consumers for anime and episode data
- **Configuration**: Environment-based config in `config/config.go`

Key data flows:
- Scraper updates anime records → anime-sync → database operations (add/update/delete)
- Supports multiple message sources: Pulsar and Kafka

## Development Commands

### Database Operations
```bash
# Run database migrations up
go run cmd/main.go migrate up

# Run database migrations down  
go run cmd/main.go migrate down

# Create new migration
make migrate-create name=<migration_name>
```

### Application Commands
```bash
# Build the application
go build -o anime-sync cmd/main.go

# Run anime sync service (Pulsar)
go run cmd/main.go serve-anime

# Run anime episode sync service (Pulsar)  
go run cmd/main.go serve-anime-episode

# Run anime sync service (Kafka)
go run cmd/main.go serve-anime-kafka

# Run anime episode sync service (Kafka)
go run cmd/main.go serve-anime-episode-kafka

# Run backwards compatibility anime service
go run cmd/main.go serve-anime-backwards-compat
```

### Development Environment
```bash
# Start MySQL and Redis containers
docker-compose up -d

# Generate mocks
make mocks
```

## Key Components

- **Processors**: Generic processor pattern (`ProcessorImpl[T]`) with retry logic and JSON parsing
- **Repositories**: GORM-based anime and episode repositories with MySQL backend
- **Event Handlers**: Handle incoming messages for anime and episode updates
- **Configuration**: Supports multiple message brokers (Pulsar/Kafka) and feature flags

## Database Schema

- **anime table**: Core anime entity storage
- **episode table**: Episode information with air dates
- Uses golang-migrate for schema versioning

## Dependencies

- **Message Brokers**: Apache Pulsar, Confluent Kafka
- **Database**: MySQL 8.0.33 with GORM
- **CLI Framework**: Cobra
- **Logging**: Uber Zap
- **Configuration**: jinzhu/configor with environment variable support