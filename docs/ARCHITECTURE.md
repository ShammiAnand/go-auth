# Go-Auth Architecture Documentation

## Table of Contents

1. [System Overview](#system-overview)
2. [Architecture Diagram](#architecture-diagram)
3. [Component Breakdown](#component-breakdown)
4. [Data Flow](#data-flow)
5. [Authentication Flow](#authentication-flow)
6. [RBAC Flow](#rbac-flow)
7. [Database Schema](#database-schema)
8. [API Endpoints](#api-endpoints)
9. [Security Considerations](#security-considerations)

---

## System Overview

Go-Auth is a lightweight authentication microservice designed for multi-backend architectures. It provides:

- **JWT-based authentication** with RS256 signing
- **JWKS (JSON Web Key Set)** endpoint for public key distribution
- **Role-Based Access Control (RBAC)** with flexible permission management
- **Email verification** and password reset workflows
- **Audit logging** for all RBAC operations
- **Redis caching** for sessions and JWKS keys
- **PostgreSQL** for persistent data storage

### Design Principles

- **Stateless Authentication**: JWT tokens enable stateless authentication
- **Public Key Distribution**: JWKS allows multiple services to verify tokens independently
- **Separation of Concerns**: Modular architecture with clear boundaries
- **Idempotent Operations**: RBAC initialization can run multiple times safely
- **Audit Trail**: All RBAC changes are logged with actor information

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                           Go-Auth Service                            │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                     HTTP Server (Gin)                        │   │
│  │                                                               │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │   │
│  │  │ Middleware   │  │ Middleware   │  │ Middleware   │     │   │
│  │  │ - Logger     │  │ - CORS       │  │ - Auth       │     │   │
│  │  │ - RequestID  │  │ - Recovery   │  │ - Permissions│     │   │
│  │  └──────────────┘  └──────────────┘  └──────────────┘     │   │
│  │                                                               │   │
│  │  ┌────────────────────────────────────────────────────────┐ │   │
│  │  │                  Router Groups                         │ │   │
│  │  │                                                         │ │   │
│  │  │  /api/v1/auth/*         /api/v1/rbac/*                │ │   │
│  │  │  ┌───────────────┐      ┌───────────────┐            │ │   │
│  │  │  │ Auth Module   │      │ RBAC Module   │            │ │   │
│  │  │  │ - signup      │      │ - roles       │            │ │   │
│  │  │  │ - signin      │      │ - permissions │            │ │   │
│  │  │  │ - logout      │      │ - user roles  │            │ │   │
│  │  │  │ - verify      │      │ - audit logs  │            │ │   │
│  │  │  │ - reset pwd   │      │               │            │ │   │
│  │  │  └───────────────┘      └───────────────┘            │ │   │
│  │  └────────────────────────────────────────────────────────┘ │   │
│  │                                                               │   │
│  │  ┌────────────────────────────────────────────────────────┐ │   │
│  │  │              Service Layer                             │ │   │
│  │  │  - AuthService                                         │ │   │
│  │  │  - RBACService                                         │ │   │
│  │  │  - EmailService                                        │ │   │
│  │  │  - BootstrapService                                    │ │   │
│  │  └────────────────────────────────────────────────────────┘ │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    Cobra CLI Commands                        │   │
│  │                                                               │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │   │
│  │  │  server  │  │   init   │  │  admin   │  │   jobs   │   │   │
│  │  │          │  │          │  │          │  │          │   │   │
│  │  │ Start    │  │ Bootstrap│  │ Create   │  │ JWKS     │   │   │
│  │  │ HTTP     │  │ RBAC     │  │ Superuser│  │ Refresh  │   │   │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │   │
│  └─────────────────────────────────────────────────────────────┘   │
└───────────────────────────────┬───────────────────────────────────┘
                                 │
                 ┌───────────────┼───────────────┐
                 │               │               │
        ┌────────▼────────┐ ┌───▼────────┐ ┌───▼────────┐
        │   PostgreSQL    │ │   Redis    │ │  Mailhog   │
        │                 │ │            │ │  (Dev) /   │
        │ - Users         │ │ - Sessions │ │  SES (Prod)│
        │ - Roles         │ │ - JWKS     │ │            │
        │ - Permissions   │ │ - Cache    │ │ - Verify   │
        │ - UserRoles     │ │            │ │ - Reset    │
        │ - RolePerms     │ │            │ │ - Welcome  │
        │ - AuditLogs     │ │            │ │            │
        │ - EmailLogs     │ │            │ │            │
        └─────────────────┘ └────────────┘ └────────────┘
```

---

## Component Breakdown

### 1. HTTP Server Layer (Gin Framework)

**Location**: `cmd/server.go`

- Handles HTTP requests and responses
- Applies global middleware (logging, CORS, recovery)
- Routes requests to appropriate module handlers
- Provides health check endpoints (`/health`, `/ready`)

**Middleware**:
- **Logger**: Structured logging with request ID
- **CORS**: Cross-origin resource sharing configuration
- **RequestID**: Unique identifier for each request
- **RequireAuth**: JWT validation middleware
- **RequirePermission**: Permission-based access control

### 2. Module Layer

#### Auth Module (`internal/modules/auth/`)

**Controller** (`controller/auth_controller.go`):
- Handles HTTP requests for authentication operations
- Validates input using Gin binding
- Delegates business logic to AuthService

**Service** (`service/auth_service.go`):
- `Signup()`: Create new user account, assign default role, send verification email
- `Signin()`: Authenticate user, create JWT token, store session in Redis
- `Logout()`: Invalidate session in Redis
- `GetUserInfo()`: Retrieve user profile
- `UpdateProfile()`: Update user information
- `ForgotPassword()`: Generate reset token, send email
- `ResetPassword()`: Validate token, update password
- `VerifyEmail()`: Confirm email address
- `ResendVerification()`: Regenerate verification token

**Router** (`router.go`):
- Registers routes under `/api/v1/auth`
- Applies authentication middleware where needed

#### RBAC Module (`internal/modules/rbac/`)

**Controller** (`controller/rbac_controller.go`):
- Handles HTTP requests for RBAC operations
- Extracts actor ID from authenticated context
- Validates role and permission IDs

**Service** (`service/rbac_service.go`):
- `ListRoles()`: Get all roles
- `GetRole()`: Get role with permissions
- `ListPermissions()`: Get all permissions
- `GetUserRoles()`: Get roles assigned to user
- `AssignRole()`: Assign role to user (with max_users validation)
- `RemoveRole()`: Remove role from user
- `GetUserPermissions()`: Compute effective permissions from all roles
- `UpdateRolePermissions()`: Modify role permissions (system roles protected)
- `GetAuditLogs()`: Query audit logs with filters
- `createAuditLog()`: Internal helper to log RBAC changes

**Bootstrap** (`bootstrap/bootstrap.go`):
- `BootstrapPermissions()`: Idempotent permission creation from YAML
- `BootstrapRoles()`: Idempotent role creation with permission assignment
- Supports wildcard permission matching (`*`, `users.*`)

**Router** (`router.go`):
- Registers routes under `/api/v1/rbac`
- Public routes for listing roles/permissions
- Protected routes for management operations

#### Email Module (`internal/modules/email/`)

**Provider Interface** (`provider/provider.go`):
```go
type EmailProvider interface {
    SendEmail(message *models.EmailMessage) error
}
```

**Implementations**:
- **MailhogProvider** (`provider/mailhog.go`): SMTP for local development
- **SESProvider** (`provider/ses.go`): AWS SES stub for production

**Service** (`service/email_service.go`):
- `SendVerificationEmail()`: HTML/text email with verification link
- `SendPasswordResetEmail()`: Reset link with token
- `SendWelcomeEmail()`: Onboarding message
- `GenerateVerificationToken()`: Create email verification token
- `GeneratePasswordResetToken()`: Create password reset token
- Logs all email delivery attempts to `email_logs` table

### 3. CLI Layer (Cobra Commands)

**Root Command** (`cmd/root.go`):
- Entry point for all CLI operations
- Loads configuration

**Server Command** (`cmd/server.go`):
- Starts HTTP server on configured port
- Initializes database, Redis, JWKS keys
- Handles graceful shutdown on SIGTERM/SIGINT

**Init Command** (`cmd/init.go`):
- Bootstraps RBAC from `configs/rbac-config.yaml`
- Idempotent: can run multiple times safely
- Reports created/updated permissions and roles

**Admin Command** (`cmd/admin.go`):
- `create-superuser`: Creates user with super-admin role
- Validates super-admin role exists
- Checks max_users constraint

**Jobs Command** (`cmd/jobs.go`):
- `jwks-refresh`: Rotates JWKS keys at specified interval
- Generates new RSA key pair
- Updates Redis cache

### 4. Database Layer (Ent ORM)

**Schemas** (`ent/schema/`):
- `users.go`: User accounts with email, password hash, name fields
- `roles.go`: Roles with code, system flag, max_users constraint
- `permissions.go`: Permissions with code, resource, action
- `user_roles.go`: Join table for user-role relationships
- `role_permissions.go`: Join table for role-permission relationships
- `audit_logs.go`: RBAC change tracking
- `email_logs.go`: Email delivery tracking
- `email_verifications.go`: Email verification tokens
- `password_resets.go`: Password reset tokens

**Auto-migration**: `storage.AutoMigrate()` runs on server start

### 5. Storage Layer

**Database** (`internal/storage/postgres.go`):
- Connection pooling
- Auto-migration
- Ent client initialization

**Redis** (`internal/storage/redis.go`):
- Session storage
- JWKS caching
- Rate limiting (future)

### 6. Authentication & Security

**JWT** (`internal/auth/jwt.go`):
- `GenerateToken()`: Creates RS256-signed JWT with user claims
- `InitializeKeys()`: Generates RSA key pair, caches in Redis
- `GetPublicKeyFromCache()`: Retrieves public key for verification

**Password Hashing** (`internal/auth/passwords.go`):
- `HashPassword()`: bcrypt hashing with cost 10
- `ComparePasswords()`: Constant-time comparison

---

## Data Flow

### Signup Flow

```
1. Client → POST /api/v1/auth/signup
   {email, password, first_name, last_name}

2. AuthController → AuthService.Signup()

3. AuthService:
   - Hash password with bcrypt
   - Create user in database (ent)
   - Assign default role (from roles.is_default = true)
   - Generate verification token
   - Create email_verifications record

4. EmailService.SendVerificationEmail()
   - Generate HTML/text templates
   - Send via MailhogProvider/SESProvider
   - Log to email_logs table

5. Response → Client
   {status: "success", message: "Check email", data: {user_id, email}}
```

### Signin Flow

```
1. Client → POST /api/v1/auth/signin
   {email, password}

2. AuthController → AuthService.Signin()

3. AuthService:
   - Find user by email
   - Compare password hash (bcrypt)
   - Check is_active and is_verified
   - Generate JWT token (RS256)
   - Store session in Redis (key: "session:{user_id}", value: token, TTL: 24h)
   - Update last_login timestamp

4. Response → Client
   {status: "success", data: {token, user: {...}}}
```

### Protected Route Flow

```
1. Client → GET /api/v1/auth/me
   Headers: Authorization: Bearer <token>

2. Middleware.RequireAuth():
   - Extract token from Authorization header
   - Parse JWT and get kid (key ID)
   - Fetch public key from Redis (GetPublicKeyFromCache)
   - Verify signature
   - Validate expiration
   - Check session exists in Redis
   - Set user_id in Gin context

3. AuthController.GetUserInfo():
   - Get user_id from context
   - Query database for user
   - Return user info

4. Response → Client
   {status: "success", data: {user_id, email, first_name, last_name, ...}}
```

### Role Assignment Flow

```
1. Admin Client → POST /api/v1/rbac/users/assign-role
   Headers: Authorization: Bearer <admin_token>
   Body: {user_id, role_id}

2. Middleware.RequireAuth():
   - Validate admin's JWT
   - Set admin_id (actor_id) in context

3. RBACController.AssignRole():
   - Extract actor_id from context
   - Validate request body
   - Call RBACService.AssignRole(user_id, role_id, actor_id)

4. RBACService.AssignRole():
   - Check user exists
   - Check role exists
   - Check max_users constraint
   - Check if already assigned
   - Create user_roles record
   - Create audit_logs record (actor_id, action: "role.assign", metadata)

5. Response → Admin Client
   {status: "success", message: "Role assigned successfully"}
```

---

## Authentication Flow

```
┌──────────┐
│  Client  │
└────┬─────┘
     │
     │ 1. POST /api/v1/auth/signin
     │    {email, password}
     ▼
┌─────────────────┐
│ AuthController  │
└────┬────────────┘
     │
     │ 2. Validate input
     ▼
┌─────────────────┐
│  AuthService    │
│                 │
│  3. Find user   │────────┐
│  4. Verify pwd  │        │
│  5. Gen JWT     │        │
│  6. Store sess  │◄───┐   │
└────┬────────────┘    │   │
     │                 │   │
     │                 │   │
     ▼                 │   │
┌──────────┐     ┌─────▼───▼──┐
│  Client  │◄────│  Response  │
│          │     │  {token}   │
└────┬─────┘     └────────────┘
     │
     │ 7. Store token
     │
     │ 8. GET /api/v1/auth/me
     │    Headers: Bearer <token>
     ▼
┌──────────────────┐
│  RequireAuth     │
│  Middleware      │
│                  │
│  9. Extract JWT  │
│  10. Get pub key │───────┐
│  11. Verify sig  │       │
│  12. Check Redis │◄───┐  │
└────┬─────────────┘    │  │
     │              ┌───▼──▼───┐
     │              │  Redis   │
     ▼              └──────────┘
┌──────────────────┐
│ AuthController   │
│  GetUserInfo()   │
└────┬─────────────┘
     │
     │ 13. Get user_id from context
     │ 14. Query database
     ▼
┌──────────────────┐
│  PostgreSQL      │
└────┬─────────────┘
     │
     │ 15. User data
     ▼
┌──────────┐
│  Client  │
└──────────┘
```

---

## RBAC Flow

### Permission Computation

```
User → UserRoles → Roles → RolePermissions → Permissions

Example:
User ID: 123
  ↓
UserRoles:
  - role_id: 1 (admin)
  - role_id: 3 (editor)
  ↓
RolePermissions (role_id = 1):
  - permission_id: 10 (users.read)
  - permission_id: 11 (users.write)
  ↓
RolePermissions (role_id = 3):
  - permission_id: 20 (content.read)
  - permission_id: 21 (content.write)
  ↓
Computed Permissions (deduplicated):
  [users.read, users.write, content.read, content.write]
```

### Audit Logging

Every RBAC operation creates an audit log:

```json
{
  "id": "uuid",
  "actor_id": "admin_user_id",
  "action_type": "role.assign",
  "resource_type": "user_role",
  "resource_id": "target_user_id",
  "metadata": {
    "user_id": "target_user_id",
    "role_id": 2
  },
  "ip_address": "192.168.1.1",
  "user_agent": "curl/7.68.0",
  "created_at": "2025-10-19T10:30:00Z"
}
```

---

## Database Schema

### Core Tables

**users**
- `id` (UUID, PK)
- `email` (string, unique)
- `password_hash` (string)
- `first_name` (string)
- `last_name` (string)
- `is_active` (bool, default: true)
- `is_verified` (bool, default: false)
- `last_login` (timestamp)
- `created_at` (timestamp)
- `updated_at` (timestamp)

**roles**
- `id` (int, PK)
- `code` (string, unique)
- `name` (string)
- `description` (string, optional)
- `is_system` (bool, default: false)
- `is_default` (bool, default: false)
- `max_users` (int, optional)
- `created_at` (timestamp)
- `updated_at` (timestamp)

**permissions**
- `id` (int, PK)
- `code` (string, unique)
- `name` (string)
- `description` (string, optional)
- `resource` (string, optional)
- `action` (string, optional)
- `created_at` (timestamp)
- `updated_at` (timestamp)

### Join Tables

**user_roles**
- `id` (int, PK)
- `user_id` (UUID, FK → users)
- `role_id` (int, FK → roles)
- `assigned_at` (timestamp)
- `assigned_by` (UUID, optional FK → users)
- UNIQUE(user_id, role_id)

**role_permissions**
- `id` (int, PK)
- `role_id` (int, FK → roles)
- `permission_id` (int, FK → permissions)
- UNIQUE(role_id, permission_id)

### Audit & Email Tables

**audit_logs**
- `id` (UUID, PK)
- `actor_id` (UUID, optional FK → users)
- `action_type` (string)
- `resource_type` (string)
- `resource_id` (string, optional)
- `metadata` (JSON)
- `changes` (JSON, optional)
- `ip_address` (string, optional)
- `user_agent` (string, optional)
- `created_at` (timestamp)

**email_logs**
- `id` (UUID, PK)
- `user_id` (UUID, FK → users)
- `email_type` (string)
- `recipient` (string)
- `subject` (string)
- `status` (string: sent, failed)
- `error_message` (string, optional)
- `sent_at` (timestamp)

**email_verifications**
- `id` (UUID, PK)
- `user_id` (UUID, FK → users)
- `token` (string, unique)
- `expires_at` (timestamp)
- `verified_at` (timestamp, optional)
- `created_at` (timestamp)

**password_resets**
- `id` (UUID, PK)
- `user_id` (UUID, FK → users)
- `token` (string, unique)
- `expires_at` (timestamp)
- `used_at` (timestamp, optional)
- `created_at` (timestamp)

---

## API Endpoints

### Authentication (`/api/v1/auth`)

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/signup` | No | Create new account |
| POST | `/signin` | No | Authenticate user |
| POST | `/logout` | Yes | Invalidate session |
| GET | `/me` | Yes | Get user info |
| PUT | `/me` | Yes | Update profile |
| POST | `/forgot-password` | No | Request password reset |
| POST | `/reset-password` | No | Complete password reset |
| GET | `/verify-email` | No | Verify email address |
| POST | `/resend-verification` | No | Resend verification email |

### RBAC (`/api/v1/rbac`)

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/roles` | No | List all roles |
| GET | `/roles/:id` | No | Get role with permissions |
| GET | `/permissions` | No | List all permissions |
| GET | `/users/:user_id/roles` | Yes | Get user's roles |
| GET | `/users/:user_id/permissions` | Yes | Get computed permissions |
| POST | `/users/assign-role` | Yes | Assign role to user |
| POST | `/users/remove-role` | Yes | Remove role from user |
| PUT | `/roles/:id/permissions` | Yes | Update role permissions |
| GET | `/audit-logs` | Yes | Query audit logs |

### Public

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/.well-known/jwks.json` | No | JWKS public keys |
| GET | `/health` | No | Basic health check |
| GET | `/ready` | No | Readiness check |

---

## Security Considerations

### 1. Token Security

- **RS256 Algorithm**: Asymmetric signing prevents token forgery
- **JWKS Rotation**: Keys should be rotated periodically (24h interval)
- **Short Expiration**: Tokens expire after 24 hours
- **Session Invalidation**: Logout removes session from Redis

### 2. Password Security

- **bcrypt Hashing**: Cost factor 10, salt included
- **No Plain Text**: Passwords never stored or logged
- **Reset Tokens**: Single-use, time-limited (1 hour)

### 3. Email Verification

- **Required for Sensitive Operations**: Check `is_verified` before password changes
- **Token Expiration**: Verification tokens expire after 24 hours
- **One-Time Use**: Tokens invalidated after verification

### 4. RBAC Protection

- **System Roles**: Cannot be modified or deleted via API
- **Audit Logging**: All changes tracked with actor information
- **Max Users Constraint**: Prevents unlimited role assignments
- **Permission Checks**: Middleware validates permissions on protected routes

### 5. Input Validation

- **Gin Binding**: Request body validation with struct tags
- **Email Format**: Validated before user creation
- **UUID Parsing**: Prevents invalid ID attacks
- **SQL Injection**: Ent ORM provides parameterized queries

### 6. Rate Limiting (Future)

- Redis-based rate limiting per IP/user
- Prevents brute force attacks
- Configurable limits per endpoint

### 7. CORS

- Configurable allowed origins
- Credentials support for cookies
- Preflight request handling

---

## Deployment Considerations

### Prerequisites

1. **Go 1.24+**: Required for building the binary
2. **PostgreSQL 13+**: Primary database
3. **Redis 6+**: Caching and session storage
4. **SMTP Server**: Mailhog (dev) or AWS SES (prod)
5. **Environment Variables**: See `.env.sample`

### Environment Configuration

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
SECRET_KEY_ID=your-key-id
SECRET_PRIVATE_KEY=your-private-key

# API
API_PORT=42069

# Email (Production)
EMAIL_PROVIDER=ses
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=...
AWS_SECRET_ACCESS_KEY=...
```

### Startup Sequence

1. Start PostgreSQL and Redis
2. Run `go-auth init --config ./configs/rbac-config.yaml`
3. Run `go-auth admin create-superuser` (once)
4. Start JWKS refresh job: `go-auth jobs jwks-refresh --interval 24h` (background)
5. Start API server: `go-auth server --port 42069`

### Health Checks

- **Liveness**: `GET /health` (always returns 200)
- **Readiness**: `GET /ready` (checks database connectivity)

### Monitoring

- Structured JSON logging to stdout
- Request ID tracking in all logs
- Audit logs for compliance
- Email delivery logs for debugging

---

## Summary

Go-Auth provides a complete authentication and authorization solution with:

- **Modular Design**: Clear separation between auth, RBAC, and email
- **Scalability**: Stateless JWT tokens, Redis caching
- **Security**: RS256 signing, bcrypt hashing, audit logging
- **Flexibility**: YAML-based RBAC configuration, wildcard permissions
- **Developer Experience**: Cobra CLI, comprehensive documentation, Docker support

For detailed API examples and usage, see the main [README.md](../README.md).
