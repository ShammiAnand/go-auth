package api

import (
	"log"

	"github.com/shammianand/go-auth/ent"
	"github.com/shammianand/go-auth/internal/storage"
)

type APIServer struct {
	addr   string
	client *ent.Client
}

func NewAPIServer(addr string, client *ent.Client) *APIServer {
	return &APIServer{
		addr:   addr,
		client: client,
	}
}

func (s *APIServer) Run() error {

	// start with migration
	err := storage.AutoMigrate(*s.client)
	if err != nil {
		return err
	}

	log.Println("MIGRATION SUCESSFULL")
	return nil
}
