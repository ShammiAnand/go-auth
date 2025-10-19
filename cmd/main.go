package main

import (
	"flag"
	"fmt"

	"github.com/shammianand/go-auth/cmd/api"
	"github.com/shammianand/go-auth/internal/config"
	"github.com/shammianand/go-auth/internal/storage"
	"github.com/shammianand/go-auth/internal/utils"
)

var (
	migrateOnlyFlag = flag.Bool("migrate", false, "setting this true would only run the db migrations")
)

func main() {

	flag.Parse()

	entClient, err := storage.DBConnect()
	if err != nil {
		utils.Logger.Error("failed to connect to db", "error", err)
	}
	defer entClient.Close()

	if *migrateOnlyFlag {
		err = storage.AutoMigrate(*entClient)
		if err != nil {
			utils.Logger.Error("failed to migrate", "error", err)
		}

		utils.Logger.Info("migration successful")
		return
	}

	redisDB := storage.GetRedisClient()

	authServer := api.NewAPIServer(fmt.Sprintf(":%s", config.ENV_API_PORT), entClient, redisDB)
	if err := authServer.Run(); err != nil {
		utils.Logger.Error("failed to start the auth server", "error", err)
		return
	}
}
