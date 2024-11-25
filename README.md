# Chirpy

A Twitter-like HTTP server implementation in Go, built as part of the boot.dev course.

## Features

- RESTful API endpoints for chirps (tweets)
- User authentication with JWT
- Refresh token mechanism
- Database storage with PostgreSQL
- Metrics endpoint
- Webhooks integration
- File serving capabilities

## Prerequisites

- Go 1.21+
- PostgreSQL
- `goose` for migrations

## Setup

1. Clone the repository

```bash
git clone https://github.com/onkelwolle/chirpy.git
cd chirp
```

2. Install dependencies

```bash
go mod download
```

3. Configure environment

```
DB_URL="postgres://username:password@localhost:5432/chirpy?sslmode=disable"
SECRET="your-jwt-secret"
POLKA_KEY="your-webhook-secret"
```

If you want to use the /admin/reset endpoint, you need to enable dev environment:

```
PLATFORM="dev"
```

4. Run database migrations

```
goose postgres "your-db-url" up
```

## API Endpoints

### Authentication

- POST /api/users - Create new user
- POST /api/login - Login user
- POST /api/refresh - Refresh access token
- PUT /api/users - Update user

### Chirps

- GET /api/chirps - List all chirps
- GET /api/chirps/{id} - Get chirp by ID
- POST /api/chirps - Create new chirp
- DELETE /api/chirps/{id} - Delete chirp

### Metrics

- GET /admin/metrics - View metrics
- POST /admin/reset - Reset metrics

## Development

```
go build -o out && ./out
```

Server will start at http://localhost:8080
