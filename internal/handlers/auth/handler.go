package auth

import (
	"github.com/redis/go-redis/v9"
	"github.com/shammianand/go-auth/ent"
	"net/http"
)

type Handler struct {
	client *ent.Client
	cache  *redis.Client // NOTE: i am not sure if this belong here?
}

func NewHandler(client *ent.Client, cache *redis.Client) *Handler {
	return &Handler{
		client: client,
		cache:  cache,
	}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("POST /auth/login", h.handleLogin)
	router.HandleFunc("POST /auth/register", h.handleRegister)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	return
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	return
}
