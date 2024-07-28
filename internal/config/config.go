package config

import "os"

var (
	ENV_DB_USER       = os.Getenv("DB_USER")
	ENV_DB_PASS       = os.Getenv("DB_PASS")
	ENV_DB_PORT       = os.Getenv("DB_PORT")
	ENV_DB_URL        = os.Getenv("DB_URL")
	ENV_DB_NAME       = os.Getenv("DB_NAME")
	ENV_REDIS_PORT    = os.Getenv("REDIS_HOST")
	ENV_REDIS_HOST    = os.Getenv("REDIS_PORT")
	ENV_SECRET_KEY_ID = os.Getenv("SECRET_KEY_ID")
)

var (
	TokenExpiry = 100
)
