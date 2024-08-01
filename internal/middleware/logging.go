package middleware

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/shammianand/go-auth/internal/utils"
)

var log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelDebug,
}))

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &utils.WrappedWriter{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}
		next.ServeHTTP(wrapped, r)
		log.Info(
			"API Logger",
			strconv.Itoa(wrapped.StatusCode),
			r.Method,
			r.URL.Path,
			time.Since(start).Seconds(),
		)
	})
}
