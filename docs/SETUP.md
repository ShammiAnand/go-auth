# Go-Auth Setup Guide

Complete guide for setting up Go-Auth in development and production environments.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start (Development)](#quick-start-development)
3. [Detailed Setup](#detailed-setup)
4. [Configuration](#configuration)
5. [CLI Commands](#cli-commands)
6. [Docker Deployment](#docker-deployment)
7. [Production Deployment](#production-deployment)
8. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required Software

#### 1. Go 1.24 or higher

**Check version**:
```bash
go version
# Should output: go version go1.24.x ...
```

**Installation**:
- **macOS**: `brew install go`
- **Linux**: Download from [golang.org/dl](https://golang.org/dl/)
- **Windows**: Download installer from [golang.org/dl](https://golang.org/dl/)

#### 2. PostgreSQL 13 or higher

**Check version**:
```bash
psql --version
# Should output: psql (PostgreSQL) 13.x or higher
```

**Installation**:
- **macOS**: `brew install postgresql@14`
- **Linux (Ubuntu)**: `sudo apt install postgresql postgresql-contrib`
- **Windows**: Download from [postgresql.org](https://www.postgresql.org/download/)
- **Docker**: `docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=admin postgres:14`

#### 3. Redis 6 or higher

**Check version**:
```bash
redis-cli --version
# Should output: redis-cli 6.x.x or higher
```

**Installation**:
- **macOS**: `brew install redis`
- **Linux (Ubuntu)**: `sudo apt install redis-server`
- **Windows**: Use WSL or Docker
- **Docker**: `docker run -d -p 6379:6379 redis:7-alpine`

#### 4. Make (optional but recommended)

**Check**:
```bash
make --version
```

**Installation**:
- **macOS**: Included with Xcode Command Line Tools
- **Linux**: Usually pre-installed, or `sudo apt install build-essential`
- **Windows**: Install via Chocolatey `choco install make`

#### 5. Docker & Docker Compose (optional)

For containerized deployment.

**Check**:
```bash
docker --version
docker-compose --version
```

**Installation**: Follow [Docker Desktop](https://www.docker.com/products/docker-desktop/) guide

---

## Quick Start (Development)

### 1. Clone Repository

```bash
git clone <repository-url>
cd go-auth
```

### 2. Copy Environment File

```bash
cp .env.sample .env
```

Edit `.env` if needed (default values work for local development).

### 3. Start Dependencies with Docker

```bash
make dev
```

This starts:
- PostgreSQL on port 5432
- Redis on port 6379
- Mailhog SMTP on port 1025
- Mailhog UI on http://localhost:8025

### 4. Initialize RBAC

```bash
make init
```

This creates:
- Default permissions (users.read, users.write, etc.)
- Default roles (super-admin, admin, user)
- Role-permission assignments

### 5. Create Super Admin User

```bash
make create-superuser
```

Default credentials:
- Email: `admin@go-auth.local`
- Password: `SuperSecure123!`

### 6. Start API Server

```bash
make run
```

Server runs on http://localhost:42069

### 7. Test the API

```bash
# Health check
curl http://localhost:42069/health

# Signup
curl -X POST http://localhost:42069/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "first_name": "John",
    "last_name": "Doe"
  }'

# Check Mailhog for verification email
open http://localhost:8025

# Signin
curl -X POST http://localhost:42069/api/v1/auth/signin \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@go-auth.local",
    "password": "SuperSecure123!"
  }'
```

---

## Detailed Setup

### Step 1: Database Setup

#### Option A: Docker (Recommended for Development)

Already included in `docker-compose.yml`:

```yaml
postgres:
  image: postgres:14-alpine
  environment:
    POSTGRES_USER: admin
    POSTGRES_PASSWORD: admin
    POSTGRES_DB: auth
  ports:
    - "5432:5432"
```

#### Option B: Local PostgreSQL

1. **Create Database**:
   ```bash
   createdb auth
   ```

2. **Create User** (if needed):
   ```sql
   CREATE USER admin WITH PASSWORD 'admin';
   GRANT ALL PRIVILEGES ON DATABASE auth TO admin;
   ```

3. **Update `.env`**:
   ```bash
   DB_URL=localhost
   DB_PORT=5432
   DB_USER=admin
   DB_PASS=admin
   DB_NAME=auth
   ```

### Step 2: Redis Setup

#### Option A: Docker (Recommended for Development)

Already included in `docker-compose.yml`:

```yaml
redis:
  image: redis:7-alpine
  ports:
    - "6379:6379"
```

#### Option B: Local Redis

1. **Start Redis**:
   ```bash
   redis-server
   ```

2. **Verify**:
   ```bash
   redis-cli ping
   # Should return: PONG
   ```

3. **Update `.env`**:
   ```bash
   REDIS_HOST=127.0.0.1
   REDIS_PORT=6379
   ```

### Step 3: Email Setup

#### Development: Mailhog (Default)

Already included in `docker-compose.yml`:

```yaml
mailhog:
  image: mailhog/mailhog:latest
  ports:
    - "1025:1025"  # SMTP server
    - "8025:8025"  # Web UI
```

Access web UI at http://localhost:8025

#### Production: AWS SES

1. **Configure AWS Credentials**:
   ```bash
   export AWS_REGION=us-east-1
   export AWS_ACCESS_KEY_ID=your_access_key
   export AWS_SECRET_ACCESS_KEY=your_secret_key
   ```

2. **Update Email Provider** in `cmd/server.go`:
   ```go
   // Replace MailhogProvider with SESProvider
   emailProvider := provider.NewSESProvider(
       os.Getenv("AWS_REGION"),
       "noreply@yourdomain.com",
       logger,
   )
   ```

3. **Verify SES Domain**:
   - Go to AWS SES Console
   - Verify your sending domain
   - Move out of sandbox mode for production

### Step 4: Build and Install

#### Using Make (Recommended)

```bash
# Build binary
make build

# Binary location: ./bin/go-auth
./bin/go-auth --help
```

#### Manual Build

```bash
go mod tidy
go mod vendor
go build -o go-auth .
./go-auth --help
```

### Step 5: Initialize RBAC

```bash
./bin/go-auth init --config ./configs/rbac-config.yaml
```

**Output**:
```
Created 12 permissions
Updated 0 permissions
Created 3 roles: super-admin, admin, user
Updated 0 roles
RBAC initialization completed successfully
```

**Customize RBAC** by editing `configs/rbac-config.yaml`:

```yaml
permissions:
  - code: "custom.read"
    name: "Read Custom Resource"
    resource: "custom"
    action: "read"

roles:
  - code: "custom-role"
    name: "Custom Role"
    description: "Custom role with specific permissions"
    permissions:
      - "custom.read"
      - "users.read"
```

Then re-run `make init` (idempotent operation).

### Step 6: Create Super Admin

```bash
./bin/go-auth admin create-superuser \
  --email admin@yourdomain.com \
  --password YourSecurePassword123! \
  --first-name Admin \
  --last-name User
```

**Constraints**:
- Only 1 super-admin allowed (enforced by `max_users: 1` in RBAC config)
- Super-admin role must exist (created by `init` command)

### Step 7: Start JWKS Refresh Job (Background)

```bash
./bin/go-auth jobs jwks-refresh --interval 24h &
```

This rotates JWT signing keys every 24 hours for enhanced security.

### Step 8: Start API Server

```bash
./bin/go-auth server --port 42069
```

**Server Logs** (JSON format):
```json
{"time":"2025-10-19T10:00:00Z","level":"INFO","msg":"Starting HTTP server","port":"42069"}
{"time":"2025-10-19T10:00:05Z","level":"INFO","msg":"Request completed","request_id":"abc123","method":"POST","path":"/api/v1/auth/signin","status":200,"duration":"45ms"}
```

---

## Configuration

### Environment Variables

Create `.env` file in project root:

```bash
# Database Configuration
DB_URL=localhost          # PostgreSQL host
DB_PORT=5432             # PostgreSQL port
DB_USER=admin            # Database user
DB_PASS=admin            # Database password
DB_NAME=auth             # Database name

# Redis Configuration
REDIS_HOST=127.0.0.1     # Redis host
REDIS_PORT=6379          # Redis port

# JWT Configuration
SECRET_KEY_ID=key1       # Key identifier for JWKS
SECRET_PRIVATE_KEY=<RSA_PRIVATE_KEY>  # Generated by InitializeKeys

# API Configuration
API_PORT=42069           # HTTP server port

# Email Configuration (Production)
EMAIL_PROVIDER=ses       # ses or mailhog
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=...
AWS_SECRET_ACCESS_KEY=...

# Logging
LOG_LEVEL=info           # debug, info, warn, error
GIN_MODE=release         # debug or release
```

### RBAC Configuration

Edit `configs/rbac-config.yaml`:

```yaml
permissions:
  # Define permissions with code, name, resource, action
  - code: "users.read"
    name: "View Users"
    description: "Can view user information"
    resource: "users"
    action: "read"

  - code: "users.write"
    name: "Manage Users"
    description: "Can create and update users"
    resource: "users"
    action: "write"

roles:
  # Define roles with code, name, permissions
  - code: "super-admin"
    name: "Super Administrator"
    description: "Full system access"
    is_system: true        # Cannot be modified via API
    is_default: false      # Not assigned to new users
    max_users: 1          # Limit to 1 super-admin
    permissions:
      - "*"               # All permissions (wildcard)

  - code: "admin"
    name: "Administrator"
    description: "Admin with elevated privileges"
    is_system: true
    permissions:
      - "users.*"         # All user permissions
      - "rbac.*"          # All RBAC permissions

  - code: "user"
    name: "User"
    description: "Default user role"
    is_default: true       # Assigned to new signups
    permissions:
      - "users.read.self"  # Can only read own profile
      - "users.write.self" # Can only update own profile
```

**Wildcard Matching**:
- `*`: All permissions
- `users.*`: All permissions starting with "users."
- `users.read`: Exact match only

---

## CLI Commands

### Server Commands

```bash
# Start HTTP server
go-auth server [--port PORT]

# Examples:
go-auth server                  # Use port from .env
go-auth server --port 8080      # Override port
```

### Initialization Commands

```bash
# Initialize RBAC from config
go-auth init [--config PATH]

# Examples:
go-auth init                                    # Use default config
go-auth init --config /path/to/custom-rbac.yaml # Custom config
```

### Admin Commands

```bash
# Create super-admin user
go-auth admin create-superuser \
  --email EMAIL \
  --password PASSWORD \
  --first-name FIRST \
  --last-name LAST

# Example:
go-auth admin create-superuser \
  --email admin@company.com \
  --password SecurePass123! \
  --first-name Alice \
  --last-name Admin
```

### Job Commands

```bash
# Run JWKS key refresh job
go-auth jobs jwks-refresh --interval DURATION

# Examples:
go-auth jobs jwks-refresh --interval 24h     # Rotate every 24 hours
go-auth jobs jwks-refresh --interval 1h      # Rotate every hour
```

---

## Docker Deployment

### Using Docker Compose

1. **Build Image**:
   ```bash
   make docker-build
   ```

2. **Start All Services**:
   ```bash
   make docker-up
   ```

   This starts:
   - PostgreSQL
   - Redis
   - Mailhog
   - Go-Auth API server

3. **View Logs**:
   ```bash
   make docker-logs
   ```

4. **Stop Services**:
   ```bash
   make docker-down
   ```

### Custom Docker Compose

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:14-alpine
    environment:
      POSTGRES_USER: admin
      POSTGRES_PASSWORD: admin
      POSTGRES_DB: auth
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  go-auth:
    build: .
    ports:
      - "42069:42069"
    environment:
      DB_URL: postgres
      DB_PORT: 5432
      DB_USER: admin
      DB_PASS: admin
      DB_NAME: auth
      REDIS_HOST: redis
      REDIS_PORT: 6379
      API_PORT: 42069
    depends_on:
      - postgres
      - redis
    command: server

volumes:
  postgres_data:
```

---

## Production Deployment

### 1. Build for Production

```bash
# Set production environment
export GIN_MODE=release

# Build optimized binary
go build -ldflags="-s -w" -o go-auth .
```

### 2. Environment Configuration

Create production `.env`:

```bash
DB_URL=prod-postgres.example.com
DB_PORT=5432
DB_USER=auth_user
DB_PASS=<strong_password>
DB_NAME=auth_prod

REDIS_HOST=prod-redis.example.com
REDIS_PORT=6379

SECRET_KEY_ID=prod-key-2025
API_PORT=42069

EMAIL_PROVIDER=ses
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=<prod_key>
AWS_SECRET_ACCESS_KEY=<prod_secret>

LOG_LEVEL=info
GIN_MODE=release
```

### 3. Run Migrations

```bash
# Initialize RBAC (idempotent)
./go-auth init --config ./configs/rbac-config.yaml

# Create super-admin
./go-auth admin create-superuser \
  --email admin@company.com \
  --password <strong_password> \
  --first-name Admin \
  --last-name User
```

### 4. Start Services

#### Using Systemd (Recommended)

Create `/etc/systemd/system/go-auth.service`:

```ini
[Unit]
Description=Go-Auth API Server
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=go-auth
WorkingDirectory=/opt/go-auth
EnvironmentFile=/opt/go-auth/.env
ExecStart=/opt/go-auth/go-auth server
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

Create `/etc/systemd/system/go-auth-jwks.service`:

```ini
[Unit]
Description=Go-Auth JWKS Refresh Job
After=network.target redis.service

[Service]
Type=simple
User=go-auth
WorkingDirectory=/opt/go-auth
EnvironmentFile=/opt/go-auth/.env
ExecStart=/opt/go-auth/go-auth jobs jwks-refresh --interval 24h
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable go-auth go-auth-jwks
sudo systemctl start go-auth go-auth-jwks
sudo systemctl status go-auth
```

#### Using Docker

```bash
docker run -d \
  --name go-auth \
  --restart always \
  -p 42069:42069 \
  --env-file .env \
  go-auth:latest \
  server
```

### 5. Reverse Proxy (Nginx)

```nginx
server {
    listen 80;
    server_name auth.example.com;

    location / {
        proxy_pass http://localhost:42069;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Enable HTTPS with Let's Encrypt:

```bash
sudo certbot --nginx -d auth.example.com
```

### 6. Monitoring

#### Health Checks

```bash
# Liveness check (always returns 200)
curl http://localhost:42069/health

# Readiness check (checks database)
curl http://localhost:42069/ready
```

#### Log Monitoring

Logs are JSON formatted for easy parsing:

```bash
# View logs
journalctl -u go-auth -f

# Filter by level
journalctl -u go-auth | jq 'select(.level=="ERROR")'

# Monitor specific endpoint
journalctl -u go-auth | jq 'select(.path=="/api/v1/auth/signin")'
```

---

## Troubleshooting

### Issue: "Database connection failed"

**Check**:
```bash
# Test PostgreSQL connection
psql -h $DB_URL -p $DB_PORT -U $DB_USER -d $DB_NAME

# Verify credentials in .env
cat .env | grep DB_
```

**Solution**:
- Ensure PostgreSQL is running
- Check firewall rules
- Verify credentials

### Issue: "Redis connection failed"

**Check**:
```bash
# Test Redis connection
redis-cli -h $REDIS_HOST -p $REDIS_PORT ping

# Check Redis status
redis-cli info server
```

**Solution**:
- Ensure Redis is running
- Check firewall rules
- Verify Redis is not password-protected (or configure password)

### Issue: "super-admin role not found"

**Solution**:
```bash
# Re-run RBAC initialization
./go-auth init --config ./configs/rbac-config.yaml

# Verify roles created
psql -h localhost -U admin -d auth -c "SELECT * FROM roles;"
```

### Issue: "max_users limit reached"

**Error**: `"role has reached maximum users limit (1)"`

**Solution**:
- For super-admin: Only 1 allowed by design
- For custom roles: Edit `configs/rbac-config.yaml` and increase `max_users`, then re-run `init`

### Issue: "Email not sending"

**Mailhog (Development)**:
- Check Mailhog UI: http://localhost:8025
- Verify SMTP port 1025 is accessible

**SES (Production)**:
```bash
# Test SES sending
aws ses send-email \
  --from noreply@yourdomain.com \
  --to test@example.com \
  --subject "Test" \
  --text "Test email"
```

- Verify domain in SES Console
- Check SES sending limits
- Move out of SES sandbox mode

### Issue: "Token verification failed"

**Check**:
```bash
# Verify JWKS keys in Redis
redis-cli keys "auth:jwks:*"

# Get public key
redis-cli get "auth:jwks:key:key1"
```

**Solution**:
```bash
# Regenerate keys
redis-cli del "auth:jwks:key:key1"
redis-cli del "auth:jwks"

# Restart server (will regenerate keys)
systemctl restart go-auth
```

### Issue: "Permission denied" errors

**Check file permissions**:
```bash
ls -la /opt/go-auth/
```

**Solution**:
```bash
# Fix ownership
sudo chown -R go-auth:go-auth /opt/go-auth/

# Fix executable
chmod +x /opt/go-auth/go-auth
```

---

## Next Steps

After successful setup:

1. **Read Architecture Docs**: See [ARCHITECTURE.md](./ARCHITECTURE.md)
2. **Review API Flows**: See [API_FLOWS.md](./API_FLOWS.md)
3. **Test Endpoints**: Use Postman or curl to test API
4. **Customize RBAC**: Edit `configs/rbac-config.yaml` for your needs
5. **Integrate with Frontend**: Use JWT tokens from signin endpoint
6. **Monitor Logs**: Set up log aggregation (ELK, Datadog, etc.)
7. **Configure Backups**: Set up PostgreSQL backups
8. **Enable Rate Limiting**: Implement Redis-based rate limiting (future enhancement)

---

## Support

For issues, questions, or contributions:

- **GitHub Issues**: [Link to repo issues]
- **Documentation**: [Link to main README]
- **License**: MIT License
