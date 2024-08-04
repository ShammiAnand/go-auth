# `go-auth`

1. [Architecture Overview](#architecture-overview)
2. [API Endpoints](#api-endpoints)
3. [JWT and JWKS Implementation](#jwt-and-jwks-implementation)
4. [Database Design](#database-design)
5. [Email Integration](#email-integration)
6. [RBAC Considerations](#rbac-considerations)

## Architecture Overview

The authentication service is designed as a RESTful API using Go. It implements JWT-based authentication with JWKS for public key distribution. The service uses a hybrid database approach with PostgreSQL as the primary database and Redis for caching and high-performance operations.

## API Endpoints

### 1. `POST /auth/signup`

### 2. `POST /auth/login`

### 3. `POST /auth/logout`

### 4. `GET /auth/token/refresh`

### 5. `POST /auth/password/reset-request`

### 6. `POST /auth/password/reset`

### 7. `GET /auth/verify-email`

### 8. `POST /auth/resend-verification`

### 9. `GET /auth/me`

### 10. `PUT /auth/update-profile`

### 11. `GET /auth/.well-known/jwks.json`

---

## JWT and JWKS Implementation

### Signing Method

- Use RS256 (RSA Signature with SHA-256)

### JWT Structure

- Includes standard claims: `iss`, `sub`, `exp`, `iat`
- custom claims can be added in future: `roles`, `permissions`

### Token Lifetime

- Access tokens: 15-30 minutes
- Refresh tokens: Longer lived (e.g., 7 days)

### JWKS Implementation

- Rotate keys periodically (e.g., every 24 hours)
- Keep old keys valid for a grace period
- caching mechanism for JWKS

## Database Design

### Primary Database: PostgreSQL

### Secondary Database: Redis

#### Users Table

```sql
CREATE TABLE users (
  id UUID PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  last_login TIMESTAMP WITH TIME ZONE,
  is_active BOOLEAN DEFAULT true,
  email_verified BOOLEAN DEFAULT false,
  verification_token VARCHAR(255),
  verification_token_expiry TIMESTAMP WITH TIME ZONE,
  password_reset_token VARCHAR(255),
  password_reset_token_expiry TIMESTAMP WITH TIME ZONE,
  metadata JSONB
);
```

#### Roles Table

```sql
CREATE TABLE roles (
  id SERIAL PRIMARY KEY,
  name VARCHAR(50) UNIQUE NOT NULL,
  description TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

#### Permissions Table

```sql
CREATE TABLE permissions (
  id SERIAL PRIMARY KEY,
  name VARCHAR(50) UNIQUE NOT NULL,
  description TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

- relationships between user, roles and permission is maintained through edges
- ent.io is used as an ORM for this

---

## Email Integration (TODO)

### Email Service Interface

```go
type EmailService interface {
    SendVerificationEmail(to string, verificationLink string) error
    SendPasswordResetEmail(to string, resetLink string) error
    SendWelcomeEmail(to string, username string) error
}
```

### Email Provider Interface

```go
type EmailProvider interface {
    Send(email Email) error
}

type Email struct {
    To      string
    From    string
    Subject string
    Body    string
}
```

### Implementation Example (SendGrid)

```go
type sendGridProvider struct {
    client *sendgrid.Client
}

func (s *sendGridProvider) Send(email Email) error {
    message := sendgrid.NewSingleEmail(
        sendgrid.NewEmail(email.From, email.From),
        email.Subject,
        sendgrid.NewEmail(email.To, email.To),
        email.Body,
        email.Body,
    )
    _, err := s.client.Send(message)
    return err
}
```

### Email Flows

1. Signup Verification
2. Password Reset

Implement these flows in the respective API endpoints, generating tokens and sending emails as needed.

## RBAC Considerations

While not initially implemented, design the system to easily incorporate RBAC later:

1. Use the roles and permissions tables defined in the database schema
2. Implement middleware to check roles/permissions from JWT claims
3. Design API endpoints for role and permission management (to be implemented later)
