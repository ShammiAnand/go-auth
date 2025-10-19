package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"

	models "github.com/shammianand/go-auth/ent"
	"github.com/shammianand/go-auth/ent/migrate"
	"github.com/shammianand/go-auth/internal/config"
	"github.com/shammianand/go-auth/internal/utils"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func connectWithHostPort() (*sql.DB, error) {
	databaseUrl := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
		config.ENV_DB_USER,
		url.QueryEscape(config.ENV_DB_PASS),
		config.ENV_DB_URL,
		config.ENV_DB_PORT,
		config.ENV_DB_NAME,
	)
	utils.Logger.Info("databaseUrl", "url", databaseUrl)
	db, err := sql.Open("pgx", databaseUrl)
	if err != nil {
		log.Fatal("failed to connect to DB")
		return nil, err
	}
	return db, nil
}

func DBConnect() (*models.Client, error) {
	db, err := connectWithHostPort()

	if err != nil {
		return nil, err
	}
	drv := entsql.OpenDB(dialect.Postgres, db)
	return models.NewClient(models.Driver(drv)), nil
}

func AutoMigrate(db models.Client) error {
	return db.Schema.Create(context.Background(),
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true))
}
