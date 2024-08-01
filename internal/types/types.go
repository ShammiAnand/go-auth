package types

import "github.com/google/uuid"

type RegisterUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterUserResponse struct {
	Email string    `json:"email"`
	ID    uuid.UUID `json:"id"`
}

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUserResponse struct {
	Token    string `json:"token"`
	IsActive bool   `json:"is_active"`
}
