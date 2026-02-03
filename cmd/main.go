package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"subscription-service/config"
	"subscription-service/internal/repository"
	"subscription-service/internal/service"
	"subscription-service/internal/transport"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, err := sql.Open("pgx", cfg.DBUrl)
	if err != nil {
		logger.Error("failed to connect db", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if _, err := db.Exec(initSQL); err != nil {
		logger.Error("migration failed", "err", err)
	}

	repo := repository.NewRepository(db)
	svc := service.NewService(repo, logger)
	handler := transport.NewHandler(svc)

	server := &http.Server{
		Addr:    cfg.Port,
		Handler: handler.InitRoutes(),
	}

	go func() {
		logger.Info("Starting server", "addr", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "err", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "err", err)
	}

	logger.Info("Server exited")
}

const initSQL = `
CREATE TABLE IF NOT EXISTS subscriptions (
	id SERIAL PRIMARY KEY,
	service_name TEXT NOT NULL,
	price INT NOT NULL,
	user_id TEXT NOT NULL,
	start_date DATE NOT NULL,
	end_date DATE
);
`
