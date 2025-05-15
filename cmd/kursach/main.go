package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"kursach/internal/config"
	"kursach/internal/http-server/handlers"
	"kursach/internal/http-server/middleware/logger"
	"kursach/internal/logger/sl"
	"kursach/internal/storage/postgres"
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
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting server", slog.String("env", cfg.Env))
	log.Debug("debug logging enabled")

	db, err := postgres.New(cfg.Postgres)

	if err != nil {
		log.Error("failed to initialize storage", sl.Err(err))
		os.Exit(1)
	}
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(logger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	userHandler := handlers.UserHandler{UserStorage: postgres.NewUserStorage(db.DB())}
	updateUserHandler := handlers.UpdateUserHandler{UserStorage: postgres.NewUserStorage(db.DB())}
	postHandler := handlers.PostHandler{PostStorage: postgres.NewPostStorage(db.DB()), UserStorage: postgres.NewUserStorage(db.DB())}
	fs := http.FileServer(http.Dir("./uploads"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	router.Post("/register", userHandler.Register)
	router.Post("/auth", userHandler.Login)
	router.Patch("/users", updateUserHandler.ServeHTTP)

	router.Post("/posts", postHandler.AddPost)
	router.Get("/posts", postHandler.GetPostsHandler)

	router.Post("/likes", postHandler.AddLikeHandler)
	router.Delete("/likes", postHandler.RemoveLikeHandler)

	router.Post("/comments", postHandler.AddCommentHandler)
	router.Delete("/comments", postHandler.DeleteCommentHandler)

	router.Handle("/static/*", http.StripPrefix("/static/", fs))
	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}
	if err = srv.ListenAndServe(); err != nil {
		log.Error("failed to start server", sl.Err(err))
	}
	log.Info("Server stopped", slog.String("address", cfg.HTTPServer.Address))
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}
