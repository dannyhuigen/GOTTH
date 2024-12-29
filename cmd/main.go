package main

import (
	"context"
	"errors"
	"goth/internal/database"
	"goth/internal/handlers"
	"goth/internal/store"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	m "goth/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

/*
* Set to production at build time
* used to determine what assets to load
 */
var Environment = "development"

func init() {
	os.Setenv("env", Environment)

	log.Printf("Loading env file\n")
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("/etc/.env"); err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}

	// run generate script
	exec.Command("make", "tailwind-build").Run()
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	r := chi.NewRouter()

	//cfg := config.MustLoadConfig()

	//db := database.MustOpen(cfg.DatabaseName)
	//passwordhash := passwordhash.NewHPasswordHash()

	//userStore := dbstore.NewUserStore(
	//	dbstore.NewUserStoreParams{
	//		DB:           db,
	//		PasswordHash: passwordhash,
	//	},
	//)

	//sessionStore := dbstore.NewSessionStore(
	//	dbstore.NewSessionStoreParams{
	//		DB: db,
	//	},
	//)

	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	var googleUserStore = store.NewGoogleUserPostgresStore(database.GetSqlXConnection())
	var googleAuthHandler = handlers.NewGoogleAuthHandler(googleUserStore)
	var sessionMiddleware = m.NewSessionMiddleware(googleUserStore)

	r.Group(func(r chi.Router) {
		r.Use(
			middleware.Logger,
		)
		r.Get("/auth/google/start", googleAuthHandler.StartGoogleOAuth)
		r.Get("/auth/google/callback", googleAuthHandler.HandleGoogleCallback)
	})

	r.Group(func(r chi.Router) {
		r.Use(
			middleware.Logger,
			sessionMiddleware.AddUserToContextMiddleware,
		)
		r.Get("/", handlers.NewHomeHandler().ServeHTTP)
		r.Get("/login", handlers.NewLoginHandler().ServeHTTP)
	})

	killSig := make(chan os.Signal, 1)

	signal.Notify(killSig, os.Interrupt, syscall.SIGTERM)

	srv := &http.Server{
		Addr:    ":2626",
		Handler: r,
	}

	go func() {
		// Start the server with HTTPS (using a self-signed cert here)
		err := srv.ListenAndServeTLS("server.crt", "server.key")
		if errors.Is(err, http.ErrServerClosed) {
			// If the server was closed gracefully, log the info message
			logger.Info("Server shutdown complete")
		} else if err != nil {
			// Log other errors (e.g., if the server fails to start)
			logger.Error("Server error", slog.Any("err", err))
			os.Exit(1)
		}
	}()

	logger.Info("Server started", slog.String("port", os.Getenv("FRONTEND_PORT")), slog.String("env", Environment))
	<-killSig
	logger.Info("Shutting down server")

	// Create a context with a timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to gracefully shut down the server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown failed", slog.Any("err", err))
		os.Exit(1)
	}

	logger.Info("Server shutdown complete")
}
