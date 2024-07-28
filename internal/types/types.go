package types

import "github.com/google/uuid"

type RegisterUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterUserResponse struct {
	Email string    `json:"email"`
	ID    uuid.UUID `json:"id"`
	Token string    `json:"token"`
}
