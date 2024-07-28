package main

import (
	"log"

	"github.com/shammianand/go-auth/cmd/api"
	"github.com/shammianand/go-auth/internal/storage"
)

func main() {
	entClient, err := storage.DBConnect()
	if err != nil {
		log.Fatal(err)
	}

	redisDB := storage.GetRedisClient()

	authServer := api.NewAPIServer(":42069", entClient, redisDB)
	if err := authServer.Run(); err != nil {
		log.Fatal(err)
	}
}
