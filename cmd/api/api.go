package api

import (
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"
	"github.com/shammianand/go-auth/ent"
	a "github.com/shammianand/go-auth/internal/auth"
	"github.com/shammianand/go-auth/internal/handlers/auth"
	"github.com/shammianand/go-auth/internal/middleware"
	"github.com/shammianand/go-auth/internal/storage"
	"github.com/shammianand/go-auth/internal/utils"
)

type APIServer struct {
	addr   string
	client *ent.Client
	cache  *redis.Client // NOTE: i am not sure if this belong here?
}

func NewAPIServer(addr string, client *ent.Client, cache *redis.Client) *APIServer {
	return &APIServer{
		addr:   addr,
		client: client,
		cache:  cache,
	}
}

func (s *APIServer) Run() error {

	// start with migration
	err := storage.AutoMigrate(*s.client)
	if err != nil {
		return err
	}

	router := http.NewServeMux()
	subrouter := utils.Subrouter(router, "/api/v1")

	err = a.InitializeKeys()
	if err != nil {
		log.Fatal(err)
	}

	authService := auth.NewHandler(s.client, s.cache)
	authService.RegisterRoutes(subrouter)

	log.Println("Auth Server Listening on: ", s.addr)
	return http.ListenAndServe(s.addr, middleware.Logging(router))
}
