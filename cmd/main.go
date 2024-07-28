package main

import (
	"log"

	"github.com/shammianand/go-auth/cmd/api"
	"github.com/shammianand/go-auth/internal/storage"
)

func main() {
	client, err := storage.DBConnect()
	if err != nil {
		log.Fatal(err)
	}

	authServer := api.NewAPIServer(":42069", client)
	if err := authServer.Run(); err != nil {
		log.Fatal(err)
	}
}
