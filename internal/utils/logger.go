package utils

import (
	"log/slog"
	"os"
)

// structured logger using log/slog package with DEBUG level
var Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelDebug,
}))
