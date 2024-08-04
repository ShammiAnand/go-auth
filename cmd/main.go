package main

import (
	"github.com/shammianand/go-auth/cmd/api"
	"github.com/shammianand/go-auth/internal/storage"
	"github.com/shammianand/go-auth/internal/utils"
)

func main() {

	entClient, err := storage.DBConnect()
	if err != nil {
		utils.Logger.Error("failed to connect to db", "error", err)
	}

	redisDB := storage.GetRedisClient()

	// TODO: move the addr to env
	authServer := api.NewAPIServer(":42069", entClient, redisDB)
	if err := authServer.Run(); err != nil {
		utils.Logger.Error("failed to start the auth server", "error", err)
		return
	}
}
