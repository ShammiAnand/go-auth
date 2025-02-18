package auth

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"context"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/shammianand/go-auth/ent"
	"github.com/shammianand/go-auth/ent/users"
	"github.com/shammianand/go-auth/internal/auth"
	"github.com/shammianand/go-auth/internal/types"
	"github.com/shammianand/go-auth/internal/utils"
)

type Handler struct {
	client *ent.Client
	cache  *redis.Client
	ctx    context.Context
	logger *slog.Logger

	// TODO: add an email client
}

// It creates a new Auth Handler with a Background context
func NewHandler(client *ent.Client, cache *redis.Client) *Handler {
	return &Handler{
		client: client,
		cache:  cache,
		ctx:    context.Background(),
		logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {

	// Un-Authenticated Routes
	router.HandleFunc("GET /.well-known/jwks.json", auth.JWKSHandler(h.cache))
	router.HandleFunc("POST /auth/login", h.handleLogin)
	router.HandleFunc("POST /auth/signup", h.handleRegister)

	// Authenticated Routes
	router.HandleFunc("GET /auth/refresh", auth.RefreshToken(h.cache))
	router.HandleFunc("GET /auth/me", auth.WithJWTAuth(h.handleGetMe, h.cache))
}

// requires an auth token
func (h *Handler) handleGetMe(w http.ResponseWriter, r *http.Request) {
	userId := auth.GetUserIdFromContext(r.Context())
	// get the info regarding this user from db

	uuidUserId, err := uuid.Parse(userId)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, err)
		return
	}

	user, err := h.client.Users.Query().Where(
		users.IDEQ(uuidUserId),
	).Only(h.ctx)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":         user.ID,
		"email":      user.Email,
		"created_at": user.CreatedAt.Local(),
		"updated_at": user.UpdatedAt.Local(),
		"last_login": user.LastLogin.Local(),
	})
}

// creates a new user entry in postgres,
// sends a verification email
// upon verification user can login to get the token?
// for now allowing to login without email verification
func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var payload types.RegisterUserRequest

	err := utils.ParseJSON(r, &payload)
	if err != nil {
		utils.WriteError(w, http.StatusFailedDependency, err)
		return
	}

	hashedPassword, err := auth.HashPasswords(payload.Password)

	user, err := h.client.Users.
		Create().
		SetEmail(payload.Email).
		SetPasswordHash(hashedPassword).
		Save(h.ctx)

	if err != nil {
		utils.WriteError(w, http.StatusFailedDependency, err)
		return
	}

	// TODO: send email verification to mark is_active field

	utils.WriteJSON(w, http.StatusCreated, types.RegisterUserResponse{
		ID:    user.ID,
		Email: user.Email,
	})
}

// this will verify the hashed password and generate a token
func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var payload types.LoginUserRequest

	err := utils.ParseJSON(r, &payload)
	if err != nil {
		utils.WriteError(w, http.StatusFailedDependency, err)
		return
	}

	user, err := h.client.Users.Query().Where(
		users.EmailEQ(payload.Email),
	).First(h.ctx)
	if err != nil {
		utils.WriteError(w, http.StatusFailedDependency, err)
		return
	}

	if !auth.ComparePasswords(user.PasswordHash, []byte(payload.Password)) {
		utils.WriteError(w, http.StatusMethodNotAllowed, fmt.Errorf("password does not match the email"))
		return
	}

	// generate token
	tokenString, err := auth.CreateJWT(user.ID, h.cache)
	if err != nil {
		utils.WriteError(w, http.StatusFailedDependency, err)
		return
	}

	err = h.cache.Set(h.ctx, fmt.Sprintf("access_token:%v", user.ID.String()), tokenString, 24*60*time.Minute).Err()
	if err != nil {
		h.logger.Info("unable to store to redis", "error", err)
	}

	utils.WriteJSON(w, http.StatusOK, types.LoginUserResponse{
		Token:    tokenString,
		IsActive: user.IsActive,
	})
}
