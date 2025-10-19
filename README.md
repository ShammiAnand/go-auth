# Go-Auth

> Lightweight authentication microservice with JWKS support for multi-backend architectures

![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)

## Table of Contents

1. [Overview](#overview)
2. [Features](#features)
3. [Quick Start](#quick-start)
4. [Architecture](#architecture)
5. [API Documentation](#api-documentation)
6. [CLI Commands](#cli-commands)
7. [Configuration](#configuration)
8. [Development](#development)
9. [Deployment](#deployment)
10. [Documentation](#documentation)
11. [License](#license)

## Overview

Go-Auth is a production-ready authentication microservice built with Go, designed to provide centralized authentication for microservice architectures. It implements JWT-based authentication with JWKS (JSON Web Key Set) for secure public key distribution, enabling multiple backend services to verify tokens without shared secrets.

## Features

- **JWT Authentication** - RS256 signed tokens with JWKS support
- **User Management** - Signup, signin, email verification, password reset
- **RBAC System** - Role-Based Access Control with flexible permissions
- **Email Integration** - Mailhog for development, SES-ready for production
- **CLI Interface** - Cobra-based CLI for all operations
- **Health Checks** - `/health` and `/ready` endpoints
- **Docker Support** - Multi-stage builds, docker-compose setup
- **Audit Logging** - Track all RBAC changes
- **Redis Caching** - Session management and JWKS caching
- **PostgreSQL** - Robust data persistence with ent ORM

## Quick Start

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- Make

### 1. Clone and Setup

```bash
git clone <repository-url>
cd go-auth
cp .env.sample .env
```

### 2. Start Dependencies

```bash
make dev
```

This starts PostgreSQL, Redis, and Mailhog.

### 3. Initialize RBAC

```bash
make init
```

This creates default roles (super-admin, admin, user) and permissions.

### 4. Create Super Admin

```bash
make create-superuser
```

### 5. Start API Server

```bash
make run
```

Server runs on `http://localhost:42069`

### 6. Test the API

```bash
# Signup
curl -X POST http://localhost:42069/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "first_name": "John",
    "last_name": "Doe"
  }'

# Signin
curl -X POST http://localhost:42069/api/v1/auth/signin \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
```

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐
│   Client    │────▶│   Go-Auth    │────▶│  PostgreSQL  │
│ (micro-svc) │     │  (API Server)│     │   (Users/    │
│             │◀────│              │◀────│    RBAC)     │
└─────────────┘     └──────────────┘     └──────────────┘
                           │
                           ├──────────────┐
                           │              │
                    ┌──────▼─────┐ ┌─────▼─────┐
                    │   Redis    │ │  Mailhog  │
                    │  (Cache/   │ │  (Emails) │
                    │  Sessions) │ │           │
                    └────────────┘ └───────────┘
```

### Components

- **Gin Framework** - High-performance HTTP router
- **Ent ORM** - Type-safe database queries
- **Cobra CLI** - Command-line interface
- **Redis** - Session storage, JWKS caching, rate limiting
- **PostgreSQL** - Primary data store
- **Mailhog** - Email testing (development)

## API Documentation

### Authentication Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/v1/auth/signup` | Create new account | No |
| POST | `/api/v1/auth/signin` | Authenticate user | No |
| POST | `/api/v1/auth/logout` | Invalidate session | Yes |
| GET | `/api/v1/auth/me` | Get user info | Yes |
| PUT | `/api/v1/auth/me` | Update profile | Yes |
| POST | `/api/v1/auth/forgot-password` | Request reset | No |
| POST | `/api/v1/auth/reset-password` | Complete reset | No |
| GET | `/api/v1/auth/verify-email` | Verify email | No |
| POST | `/api/v1/auth/resend-verification` | Resend email | No |
| GET | `/api/v1/.well-known/jwks.json` | Public keys | No |

### Health Check Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Basic health check |
| GET | `/ready` | Readiness check (DB connectivity) |

## CLI Commands

```bash
# Server
go-auth server [--port PORT]             # Start HTTP server

# RBAC Initialization
go-auth init [--config PATH]             # Bootstrap roles/permissions

# Admin
go-auth admin create-superuser \         # Create super-admin
  --email EMAIL \
  --password PASSWORD \
  --first-name FIRST \
  --last-name LAST

# Jobs
go-auth jobs jwks-refresh \              # JWKS key rotation job
  --interval DURATION
```

## Configuration

### Environment Variables

Create `.env` file:

```bash
# Database
DB_URL=localhost
DB_PORT=5432
DB_USER=admin
DB_PASS=admin
DB_NAME=auth

# Redis
REDIS_HOST=127.0.0.1
REDIS_PORT=6379

# JWT
SECRET_KEY_ID=your-secret-id
SECRET_PRIVATE_KEY=your-private-key

# API
API_PORT=42069
```

### RBAC Configuration

Edit `configs/rbac-config.yaml` to customize roles and permissions:

```yaml
permissions:
  - code: "users.read"
    name: "View Users"
    resource: "users"
    action: "read"

roles:
  - code: "admin"
    name: "Administrator"
    is_system: true
    permissions:
      - "users.*"
      - "rbac.*"
```

## Development

### Makefile Commands

```bash
make help              # Show all commands
make build             # Build binary
make run               # Build and run server
make init              # Initialize RBAC
make create-superuser  # Create admin user
make test              # Run tests
make gen-ent           # Generate Ent code
make clean             # Remove build artifacts
make dev               # Start dev environment
```

### Project Structure

```
go-auth/
├── cmd/              # CLI commands (Cobra)
├── configs/          # Configuration files
├── ent/              # Ent schema & generated code
├── internal/
│   ├── auth/         # JWT & password utilities
│   ├── common/       # Shared types & middleware
│   ├── config/       # Config loading
│   ├── modules/      # Feature modules
│   │   ├── auth/     # Authentication
│   │   ├── email/    # Email service
│   │   └── rbac/     # RBAC & bootstrap
│   └── storage/      # DB connections
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── README.md
```

## Documentation

For detailed information, see the following documentation:

- **[Architecture Guide](docs/ARCHITECTURE.md)** - Complete system architecture, component breakdown, data flow diagrams, and security considerations
- **[API Flow Examples](docs/API_FLOWS.md)** - Detailed API request/response examples for all major flows (signup, signin, RBAC, etc.)
- **[Setup Guide](docs/SETUP.md)** - Prerequisites, installation steps, configuration, CLI usage, and troubleshooting

### Quick Links

- [Database Schema](docs/ARCHITECTURE.md#database-schema)
- [Authentication Flow](docs/ARCHITECTURE.md#authentication-flow)
- [RBAC Flow](docs/ARCHITECTURE.md#rbac-flow)
- [API Endpoints Reference](docs/ARCHITECTURE.md#api-endpoints)
- [Security Considerations](docs/ARCHITECTURE.md#security-considerations)
- [Production Deployment](docs/SETUP.md#production-deployment)
- [Troubleshooting](docs/SETUP.md#troubleshooting)

## License

MIT License - see [LICENSE](LICENSE) file for details
