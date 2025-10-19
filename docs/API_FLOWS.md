# API Flow Examples

This document provides detailed examples of API flows with request/response examples.

## Table of Contents

1. [User Signup Flow](#user-signup-flow)
2. [User Signin Flow](#user-signin-flow)
3. [Email Verification Flow](#email-verification-flow)
4. [Password Reset Flow](#password-reset-flow)
5. [Protected Resource Access](#protected-resource-access)
6. [Role Assignment Flow](#role-assignment-flow)
7. [Permission Computation Flow](#permission-computation-flow)
8. [Audit Log Query Flow](#audit-log-query-flow)

---

## User Signup Flow

### Step 1: Client Sends Signup Request

```bash
POST http://localhost:42069/api/v1/auth/signup
Content-Type: application/json

{
  "email": "john.doe@example.com",
  "password": "SecurePass123!",
  "first_name": "John",
  "last_name": "Doe"
}
```

### Step 2: Server Processing

1. **Controller Validation** (`auth_controller.go:Signup()`)
   - Validates JSON binding
   - Checks required fields

2. **Service Layer** (`auth_service.go:Signup()`)
   - Hashes password with bcrypt (cost 10)
   - Creates user in database
   - Assigns default role (from `roles.is_default = true`)
   - Generates verification token (UUID)
   - Creates `email_verifications` record (expires in 24h)

3. **Email Service** (`email_service.go:SendVerificationEmail()`)
   - Builds HTML email template
   - Sends via MailhogProvider (dev) or SESProvider (prod)
   - Logs to `email_logs` table

### Step 3: Server Response

```json
{
  "status": "success",
  "message": "User created successfully. Please check your email to verify your account.",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "john.doe@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "is_verified": false
  }
}
```

### Step 4: User Receives Email

```
Subject: Verify your email address

Hello John,

Welcome to Go-Auth! Please verify your email address by clicking the link below:

http://localhost:3000/verify-email?token=abcd1234-5678-90ef-ghij-klmnopqrstuv

This link expires in 24 hours.

---
Go-Auth Team
```

---

## User Signin Flow

### Step 1: Client Sends Signin Request

```bash
POST http://localhost:42069/api/v1/auth/signin
Content-Type: application/json

{
  "email": "john.doe@example.com",
  "password": "SecurePass123!"
}
```

### Step 2: Server Processing

1. **Controller Validation** (`auth_controller.go:Signin()`)
   - Validates JSON binding

2. **Service Layer** (`auth_service.go:Signin()`)
   - Finds user by email
   - Compares password hash using bcrypt
   - Checks `is_active = true` and `is_verified = true`
   - Generates JWT token:
     ```go
     Claims:
       - user_id: UUID
       - email: string
       - exp: now + 24h
       - iat: now
     Signature: RS256 with private key
     ```
   - Stores session in Redis:
     ```
     Key: "session:{user_id}"
     Value: token
     TTL: 24 hours
     ```
   - Updates `users.last_login` timestamp

### Step 3: Server Response

```json
{
  "status": "success",
  "message": "Authentication successful",
  "data": {
    "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6ImtleTEifQ...",
    "user": {
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "john.doe@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "is_active": true,
      "is_verified": true,
      "last_login": "2025-10-19T10:30:00Z"
    }
  }
}
```

### Step 4: Client Stores Token

Client stores the token in:
- **Browser**: localStorage or httpOnly cookie
- **Mobile**: Secure storage (Keychain/Keystore)

---

## Email Verification Flow

### Step 1: User Clicks Verification Link

```bash
GET http://localhost:3000/verify-email?token=abcd1234-5678-90ef-ghij-klmnopqrstuv
```

Frontend extracts token and calls backend:

```bash
GET http://localhost:42069/api/v1/auth/verify-email?token=abcd1234-5678-90ef-ghij-klmnopqrstuv
```

### Step 2: Server Processing

1. **Controller** (`auth_controller.go:VerifyEmail()`)
   - Extracts token from query param

2. **Service Layer** (`auth_service.go:VerifyEmail()`)
   - Finds `email_verifications` record by token
   - Checks token not expired (`expires_at > now`)
   - Checks token not already used (`verified_at IS NULL`)
   - Updates `users.is_verified = true`
   - Updates `email_verifications.verified_at = now`

### Step 3: Server Response

```json
{
  "status": "success",
  "message": "Email verified successfully",
  "data": null
}
```

### Error Cases

**Token Expired**:
```json
{
  "status": "failure",
  "message": "Verification failed",
  "error": {
    "error_code": "TOKEN_EXPIRED",
    "error_msg": "Verification token has expired"
  }
}
```

**Token Already Used**:
```json
{
  "status": "failure",
  "message": "Verification failed",
  "error": {
    "error_code": "TOKEN_USED",
    "error_msg": "Email already verified"
  }
}
```

---

## Password Reset Flow

### Step 1: Request Password Reset

```bash
POST http://localhost:42069/api/v1/auth/forgot-password
Content-Type: application/json

{
  "email": "john.doe@example.com"
}
```

### Step 2: Server Processing

1. **Service Layer** (`auth_service.go:ForgotPassword()`)
   - Finds user by email (fails silently if not found for security)
   - Generates reset token (UUID)
   - Creates `password_resets` record (expires in 1 hour)
   - Sends reset email

### Step 3: Server Response

```json
{
  "status": "success",
  "message": "If that email exists, a password reset link has been sent",
  "data": null
}
```

### Step 4: User Receives Email

```
Subject: Reset your password

Hello John,

We received a request to reset your password. Click the link below:

http://localhost:3000/reset-password?token=xyz789-abcd-1234-efgh-567890ijklmn

This link expires in 1 hour. If you didn't request this, ignore this email.

---
Go-Auth Team
```

### Step 5: User Submits New Password

```bash
POST http://localhost:42069/api/v1/auth/reset-password
Content-Type: application/json

{
  "token": "xyz789-abcd-1234-efgh-567890ijklmn",
  "new_password": "NewSecurePass456!"
}
```

### Step 6: Server Processing

1. **Service Layer** (`auth_service.go:ResetPassword()`)
   - Finds `password_resets` record by token
   - Checks token not expired (`expires_at > now`)
   - Checks token not already used (`used_at IS NULL`)
   - Hashes new password
   - Updates `users.password_hash`
   - Updates `password_resets.used_at = now`
   - Invalidates all user sessions in Redis

### Step 7: Server Response

```json
{
  "status": "success",
  "message": "Password reset successfully. Please sign in with your new password.",
  "data": null
}
```

---

## Protected Resource Access

### Step 1: Client Requests Protected Resource

```bash
GET http://localhost:42069/api/v1/auth/me
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6ImtleTEifQ...
```

### Step 2: Middleware Processing

1. **RequireAuth Middleware** (`middleware/auth.go:RequireAuth()`)

   a. Extract token from `Authorization: Bearer <token>` header

   b. Parse JWT header to get `kid` (key ID)

   c. Fetch public key from Redis:
      ```go
      key := fmt.Sprintf("auth:jwks:key:%s", kid)
      publicKey, err := redis.Get(ctx, key).Result()
      ```

   d. Verify JWT signature using public key:
      ```go
      token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
          return publicKey, nil
      })
      ```

   e. Validate claims:
      - Check `exp` (expiration)
      - Extract `user_id`

   f. Check session exists in Redis:
      ```go
      sessionKey := fmt.Sprintf("session:%s", userID)
      exists, err := redis.Exists(ctx, sessionKey).Result()
      ```

   g. Store `user_id` in Gin context:
      ```go
      c.Set(middleware.UserIDKey, userID)
      c.Next()
      ```

### Step 3: Controller Processing

```go
func (c *AuthController) GetUserInfo(ctx *gin.Context) {
    // Get user_id from context (set by middleware)
    userID, _ := ctx.Get(middleware.UserIDKey)
    userUUID := userID.(uuid.UUID)

    // Query database
    user, err := c.service.GetUserInfo(ctx.Request.Context(), userUUID)

    // Return response
    utils.RespondSuccess(ctx, types.HTTP.Ok, "User info retrieved", user)
}
```

### Step 4: Server Response

```json
{
  "status": "success",
  "message": "User info retrieved",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "john.doe@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "is_active": true,
    "is_verified": true,
    "last_login": "2025-10-19T10:30:00Z",
    "created_at": "2025-10-15T08:00:00Z"
  }
}
```

### Error Cases

**Missing Token**:
```json
{
  "status": "failure",
  "message": "Authentication required",
  "error": {
    "error_code": "UNAUTHORIZED",
    "error_msg": "Missing or invalid authorization header"
  }
}
```

**Invalid Token**:
```json
{
  "status": "failure",
  "message": "Authentication failed",
  "error": {
    "error_code": "INVALID_TOKEN",
    "error_msg": "Token signature is invalid"
  }
}
```

**Expired Token**:
```json
{
  "status": "failure",
  "message": "Authentication failed",
  "error": {
    "error_code": "TOKEN_EXPIRED",
    "error_msg": "Token has expired"
  }
}
```

**Session Not Found**:
```json
{
  "status": "failure",
  "message": "Session invalid",
  "error": {
    "error_code": "SESSION_NOT_FOUND",
    "error_msg": "Please sign in again"
  }
}
```

---

## Role Assignment Flow

### Step 1: Admin Assigns Role to User

```bash
POST http://localhost:42069/api/v1/rbac/users/assign-role
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "role_id": 2
}
```

### Step 2: Server Processing

1. **Middleware** (`middleware/auth.go:RequireAuth()`)
   - Validates admin's JWT
   - Sets `admin_id` in context

2. **Controller** (`rbac_controller.go:AssignRole()`)
   - Validates request body
   - Extracts `actor_id` (admin_id) from context
   - Calls `RBACService.AssignRole()`

3. **Service Layer** (`rbac_service.go:AssignRole()`)

   a. Check user exists:
      ```go
      exists, err := client.Users.Query().
          Where(users.IDEQ(userID)).
          Exist(ctx)
      ```

   b. Check role exists and get details:
      ```go
      role, err := client.Roles.Get(ctx, roleID)
      ```

   c. Check `max_users` constraint:
      ```go
      if role.MaxUsers != nil {
          count, err := client.UserRoles.Query().
              Where(userroles.RoleIDEQ(roleID)).
              Count(ctx)
          if count >= *role.MaxUsers {
              return fmt.Errorf("role has reached maximum users limit")
          }
      }
      ```

   d. Check not already assigned:
      ```go
      exists, err := client.UserRoles.Query().
          Where(
              userroles.UserIDEQ(userID),
              userroles.RoleIDEQ(roleID),
          ).
          Exist(ctx)
      if exists {
          return fmt.Errorf("role already assigned")
      }
      ```

   e. Create assignment:
      ```go
      _, err = client.UserRoles.Create().
          SetUserID(userID).
          SetRoleID(roleID).
          SetNillableAssignedBy(&actorID).
          Save(ctx)
      ```

   f. Create audit log:
      ```go
      client.AuditLogs.Create().
          SetActorID(actorID).
          SetActionType("role.assign").
          SetResourceType("user_role").
          SetResourceID(userID.String()).
          SetMetadata(map[string]interface{}{
              "user_id": userID.String(),
              "role_id": roleID,
          }).
          Save(ctx)
      ```

### Step 3: Server Response

```json
{
  "status": "success",
  "message": "Role assigned successfully",
  "data": null
}
```

### Step 4: Verify Assignment

```bash
GET http://localhost:42069/api/v1/rbac/users/550e8400-e29b-41d4-a716-446655440000/roles
Authorization: Bearer <token>
```

Response:
```json
{
  "status": "success",
  "message": "User roles retrieved successfully",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "john.doe@example.com",
    "roles": [
      {
        "id": 1,
        "code": "user",
        "name": "User",
        "description": "Default user role",
        "is_system": false,
        "is_default": true
      },
      {
        "id": 2,
        "code": "admin",
        "name": "Administrator",
        "description": "Admin role with elevated privileges",
        "is_system": true,
        "is_default": false
      }
    ],
    "assigned_at": "2025-10-19T11:00:00Z"
  }
}
```

---

## Permission Computation Flow

### Step 1: Request User Permissions

```bash
GET http://localhost:42069/api/v1/rbac/users/550e8400-e29b-41d4-a716-446655440000/permissions
Authorization: Bearer <token>
```

### Step 2: Server Processing

1. **Service Layer** (`rbac_service.go:GetUserPermissions()`)

   a. Get all user roles:
      ```go
      userRolesList, err := client.UserRoles.Query().
          Where(userroles.UserIDEQ(userID)).
          All(ctx)

      roleIDs := []int{} // [1, 2]
      for _, ur := range userRolesList {
          roleIDs = append(roleIDs, ur.RoleID)
      }
      ```

   b. Get all permissions for these roles:
      ```go
      rolePerms, err := client.RolePermissions.Query().
          Where(rolepermissions.RoleIDIn(roleIDs...)).
          WithPermission().
          All(ctx)
      ```

   c. Deduplicate permissions:
      ```go
      permMap := make(map[int]*ent.Permissions)
      for _, rp := range rolePerms {
          if rp.Edges.Permission != nil {
              permMap[rp.PermissionID] = rp.Edges.Permission
          }
      }
      ```

   d. Convert to response format

### Step 3: Server Response

```json
{
  "status": "success",
  "message": "User permissions retrieved successfully",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "permissions": [
      {
        "id": 10,
        "code": "users.read",
        "name": "View Users",
        "description": "Can view user information",
        "resource": "users",
        "action": "read"
      },
      {
        "id": 11,
        "code": "users.write",
        "name": "Manage Users",
        "description": "Can create and update users",
        "resource": "users",
        "action": "write"
      },
      {
        "id": 20,
        "code": "rbac.read",
        "name": "View RBAC",
        "description": "Can view roles and permissions",
        "resource": "rbac",
        "action": "read"
      },
      {
        "id": 21,
        "code": "rbac.write",
        "name": "Manage RBAC",
        "description": "Can assign roles and modify permissions",
        "resource": "rbac",
        "action": "write"
      }
    ]
  }
}
```

### Permission Inheritance Example

```
User: john.doe@example.com
  ↓
Roles:
  1. user (is_default: true)
     - users.read.self
     - users.write.self

  2. admin (assigned by super-admin)
     - users.*  (wildcard: users.read, users.write, users.delete)
     - rbac.*   (wildcard: rbac.read, rbac.write)
  ↓
Computed Permissions (deduplicated):
  - users.read.self
  - users.write.self
  - users.read
  - users.write
  - users.delete
  - rbac.read
  - rbac.write
```

---

## Audit Log Query Flow

### Step 1: Request Audit Logs with Filters

```bash
GET http://localhost:42069/api/v1/rbac/audit-logs?actor_id=123e4567-e89b-12d3-a456-426614174000&action_type=role.assign&limit=10&offset=0
Authorization: Bearer <admin_token>
```

### Step 2: Server Processing

1. **Controller** (`rbac_controller.go:GetAuditLogs()`)
   - Binds query parameters to `AuditLogFilter` struct
   - Calls `RBACService.GetAuditLogs()`

2. **Service Layer** (`rbac_service.go:GetAuditLogs()`)

   a. Build query with filters:
      ```go
      query := client.AuditLogs.Query()

      if filter.ActorID != "" {
          actorUUID, _ := uuid.Parse(filter.ActorID)
          query = query.Where(auditlogs.ActorIDEQ(actorUUID))
      }

      if filter.ActionType != "" {
          query = query.Where(auditlogs.ActionTypeEQ(filter.ActionType))
      }

      if filter.ResourceType != "" {
          query = query.Where(auditlogs.ResourceTypeEQ(filter.ResourceType))
      }
      ```

   b. Apply pagination:
      ```go
      if filter.Limit == 0 {
          filter.Limit = 50
      }
      if filter.Limit > 100 {
          filter.Limit = 100
      }

      logs, err := query.
          Limit(filter.Limit).
          Offset(filter.Offset).
          Order(ent.Desc(auditlogs.FieldCreatedAt)).
          All(ctx)
      ```

### Step 3: Server Response

```json
{
  "status": "success",
  "message": "Audit logs retrieved successfully",
  "data": [
    {
      "id": "a1b2c3d4-e5f6-4789-a012-3456789abcde",
      "actor_id": "123e4567-e89b-12d3-a456-426614174000",
      "action_type": "role.assign",
      "resource_type": "user_role",
      "resource_id": "550e8400-e29b-41d4-a716-446655440000",
      "metadata": {
        "user_id": "550e8400-e29b-41d4-a716-446655440000",
        "role_id": 2
      },
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0 ...",
      "created_at": "2025-10-19T11:00:00Z"
    },
    {
      "id": "b2c3d4e5-f6a7-5890-b123-456789abcdef",
      "actor_id": "123e4567-e89b-12d3-a456-426614174000",
      "action_type": "role.assign",
      "resource_type": "user_role",
      "resource_id": "660f9511-f30c-52e5-b827-557766551111",
      "metadata": {
        "user_id": "660f9511-f30c-52e5-b827-557766551111",
        "role_id": 3
      },
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0 ...",
      "created_at": "2025-10-19T10:45:00Z"
    }
  ]
}
```

### Common Action Types

- `role.assign` - Role assigned to user
- `role.remove` - Role removed from user
- `role.permissions.update` - Role permissions modified
- `user.create` - User account created
- `user.update` - User profile updated
- `user.delete` - User account deleted

---

## Summary

This document covers the major API flows in Go-Auth:

1. **Signup**: User creation, role assignment, email verification
2. **Signin**: Authentication, JWT generation, session storage
3. **Email Verification**: Token-based email confirmation
4. **Password Reset**: Secure password recovery flow
5. **Protected Access**: JWT validation, session checking
6. **Role Assignment**: RBAC operations with audit logging
7. **Permission Computation**: Effective permissions from multiple roles
8. **Audit Logs**: Querying RBAC change history

All flows follow the standard `ApiResponse` format with proper error handling and security validations.
