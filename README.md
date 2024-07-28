# `go-auth` by Shammi Anand

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [API Endpoints](#api-endpoints)
3. [JWT and JWKS Implementation](#jwt-and-jwks-implementation)
4. [Database Design](#database-design)
5. [Email Integration](#email-integration)
6. [RBAC Considerations](#rbac-considerations)

## Architecture Overview

The authentication service is designed as a RESTful API using Go. It implements JWT-based authentication with JWKS for public key distribution. The service uses a hybrid database approach with PostgreSQL as the primary database and Redis for caching and high-performance operations.

## API Endpoints

### 1. `/auth/signup` (POST)

- **Purpose**: Register a new user
- **Input**: Email, password, optional user details
- **Process**:
  - Validate input
  - Check for existing users
  - Hash password using bcrypt
  - Generate and store email verification token
  - Trigger verification email
- **Response**: User ID and message (do not auto-login)

### 2. `/auth/login` (POST)

- **Purpose**: Authenticate a user
- **Input**: Email, password
- **Process**:
  - Validate credentials
  - Check if email is verified
  - Generate access and refresh tokens
- **Response**: Access token, refresh token, user info

### 3. `/auth/logout` (POST)

- **Purpose**: Log out a user
- **Input**: Refresh token
- **Process**:
  - Invalidate refresh token
  - Optionally blacklist access token
- **Response**: Success message

### 4. `/auth/token/refresh` (POST)

- **Purpose**: Obtain a new access token
- **Input**: Refresh token
- **Process**:
  - Validate refresh token
  - Generate new access token
  - Optionally rotate refresh token
- **Response**: New access token, optionally new refresh token

### 5. `/auth/password/reset-request` (POST)

- **Purpose**: Request a password reset
- **Input**: Email
- **Process**:
  - Generate and store password reset token
  - Trigger reset email
- **Response**: Success message

### 6. `/auth/password/reset` (POST)

- **Purpose**: Reset password
- **Input**: Reset token, new password
- **Process**:
  - Validate reset token
  - Update password
  - Invalidate all existing sessions for the user
- **Response**: Success message

### 7. `/auth/verify-email` (GET)

- **Purpose**: Verify user's email
- **Input**: Verification token (as query parameter)
- **Process**:
  - Validate email verification token
  - Mark email as verified
- **Response**: Success message or redirect

### 8. `/auth/resend-verification` (POST)

- **Purpose**: Resend verification email
- **Input**: Email
- **Process**:
  - Generate new verification token
  - Trigger verification email
- **Response**: Success message

### 9. `/auth/me` (GET)

- **Purpose**: Get current user's details
- **Input**: Access token (in Authorization header)
- **Process**: Retrieve user details from database
- **Response**: User details

### 10. `/auth/update-profile` (PUT)

- **Purpose**: Update user profile
- **Input**: Access token, updated profile information
- **Process**: Update user information in database
- **Response**: Updated user details

### 11. `/auth/.well-known/jwks.json` (GET)

- **Purpose**: Provide JWKS (JSON Web Key Set)
- **Process**: Return current public keys used for token verification
- **Response**: JWKS

## JWT and JWKS Implementation

### Signing Method

- Use RS256 (RSA Signature with SHA-256)

### JWT Structure

- Include standard claims: `iss`, `sub`, `exp`, `iat`
- Add custom claims: `roles`, `permissions`

### Token Lifetime

- Access tokens: 15-30 minutes
- Refresh tokens: Longer lived (e.g., 7 days)

### JWKS Implementation

- Rotate keys periodically (e.g., every 24 hours)
- Keep old keys valid for a grace period
- Implement caching mechanism for JWKS

## Database Design

### Primary Database: PostgreSQL

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

#### User_Roles Table

```sql
CREATE TABLE user_roles (
  user_id UUID REFERENCES users(id),
  role_id INTEGER REFERENCES roles(id),
  PRIMARY KEY (user_id, role_id)
);
```

#### Role_Permissions Table

```sql
CREATE TABLE role_permissions (
  role_id INTEGER REFERENCES roles(id),
  permission_id INTEGER REFERENCES permissions(id),
  PRIMARY KEY (role_id, permission_id)
);
```

### Secondary Database: Redis

#### Data Structures

1. Access Tokens:

   - Key: `access_token:{token_id}`
   - Value: JSON string containing token data
   - Expiration: Set to token expiry time

2. Refresh Tokens:

   - Key: `refresh_token:{token_id}`
   - Value: JSON string containing token data
   - Expiration: Set to refresh token expiry time

3. User Sessions:

   - Key: `user_session:{user_id}`
   - Value: Hash containing session data
   - Expiration: Set to session timeout

4. Rate Limiting:
   - Key: `rate_limit:{ip_address}`
   - Value: Sorted set of timestamp:count pairs
   - Expiration: Set based on rate limit window

## Email Integration

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
