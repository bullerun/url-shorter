package main

import (
	"REST-API-Service/internal/config"
	"REST-API-Service/internal/http-server/handlers/url/save"
	mwlogger "REST-API-Service/internal/http-server/middleware/logger"
	"REST-API-Service/internal/lib/logger/handlers/slogpretty"
	"REST-API-Service/internal/lib/logger/sl"
	"REST-API-Service/internal/storage/postgreSQL"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	//TODO: init config
	cfg := config.MustLoad()

	//TODO: init logger
	log := setupLog(cfg.Env)

	//TODO: init storage
	storage, err := postgreSQL.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to initial storage", sl.Err(err))
		os.Exit(1)
	}

	//TODO: init router
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(mwlogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Post("/url", save.New(log, storage))

	//TODO: run server
	log.Info("server starting", slog.String("address", cfg.Address))
	server := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}
	err = server.ListenAndServe()
	if err != nil {
		log.Error("failed to start server", sl.Err(err))
	}
	log.Error("server stopped")
}
func setupLog(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
		//log = setupPrettySlog()
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}
func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
