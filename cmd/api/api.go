package api

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/shammianand/go-auth/ent"
	a "github.com/shammianand/go-auth/internal/auth"
	"github.com/shammianand/go-auth/internal/handlers/auth"
	"github.com/shammianand/go-auth/internal/middleware"
	"github.com/shammianand/go-auth/internal/utils"
)

type APIServer struct {
	addr string

	// TODO: interface it out with a db pool
	client *ent.Client

	cache  *redis.Client
	logger *slog.Logger
}

func NewAPIServer(addr string, client *ent.Client, cache *redis.Client) *APIServer {
	return &APIServer{
		addr:   addr,
		client: client,
		cache:  cache,
		logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}
}

func (s *APIServer) Run() error {

	router := http.NewServeMux()
	subrouter := utils.Subrouter(router, "/api/v1")

	// TODO: later on we need to think about rotating these keys
	// at least every 24 hours
	// maybe have a job that calls this every 24 hours
	err := a.InitializeKeys(s.cache)
	if err != nil {
		s.logger.Error("failed to InitializeKeys", "error", err)
		return err
	}

	authService := auth.NewHandler(s.client, s.cache)
	authService.RegisterRoutes(subrouter)

	s.logger.Info("Auth Server Started", "addr", s.addr)
	return http.ListenAndServe(s.addr, middleware.Logging(router))
}
