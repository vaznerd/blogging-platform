package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"codeberg.org/vaznerd/blogging-platform/internal/auth"
	"codeberg.org/vaznerd/blogging-platform/internal/config"
	"codeberg.org/vaznerd/blogging-platform/internal/logger"
	"codeberg.org/vaznerd/blogging-platform/internal/server"
	"codeberg.org/vaznerd/blogging-platform/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/resend/resend-go/v3"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	bootstrapLogger := slog.Default()
	bootstrapLogger.Info("Starting Go Blogging platform API...")
	cfg, err := config.LoadConfig()
	if err != nil {
		bootstrapLogger.Error("Failed to load configuration", "error", err)
		return err
	}
	if err := cfg.Validate(); err != nil {
		bootstrapLogger.Error("Configuration validation failed", "error", err)
		return err
	}
	log := logger.NewLogger(&cfg.Log)
	cfg.LogAllConfig(log)

	mail := resend.NewClient(cfg.Resend.APIKey)

	q := url.Values{}
	q.Set("sslmode", cfg.DB.SSLMode)
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.DB.User, cfg.DB.Password),
		Host:     net.JoinHostPort(cfg.DB.HostGo, strconv.Itoa(cfg.DB.Port)),
		Path:     cfg.DB.Name,
		RawQuery: q.Encode(),
	}
	dbconfig, err := pgxpool.ParseConfig(u.String())
	if err != nil {
		log.Error("failed to parse Postgress config", "error", err)
		return err
	}
	dbconfig.MaxConns = cfg.DB.MaxOpenConns
	dbconfig.MinConns = cfg.DB.MaxIdleConns
	dbconfig.MaxConnLifetime = cfg.DB.ConnMaxLifetime
	dbconfig.MaxConnIdleTime = cfg.DB.ConnMaxIdleTime
	dbpool, err := pgxpool.NewWithConfig(context.Background(), dbconfig)
	if err != nil {
		log.Error("failed to connect to DB", "error", err)
		return err
	}
	if err = dbpool.Ping(context.Background()); err != nil {
		log.Error("failed to ping Postgress", "error", err)
		dbpool.Close()
		return err
	}
	log.Info("postgress connection established")

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err = rdb.Ping(context.Background()).Err(); err != nil {
		log.Error("failed to ping Redis", "error", err)
		if err = rdb.Close(); err != nil {
			log.Error("rdb.Close", "error", err)
		}
		return err
	}
	log.Info("redis connection established")

	authService := auth.NewService(&cfg.JWT, dbpool)
	userRepository := user.NewRepository(dbpool, rdb, log)
	userService := user.NewService(userRepository, log, mail)
	router := server.NewRouter(userService, authService, log, mail)
	server := &http.Server{
		Addr:           ":" + cfg.Server.Port,
		Handler:        router,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		IdleTimeout:    cfg.Server.IdleTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	go func() {
		log.Info("Server starting", "address", server.Addr)
		log.Info("Health check available", "url", fmt.Sprintf("http://localhost:%s/health", cfg.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	log.Info("Received shutdown signal", "signal", sig)
	log.Info("Shutting down server gracefully...")

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
		return err
	}
	dbpool.Close()
	if err := rdb.Close(); err != nil {
		log.Error("rdb.Close", "error", err)
	}

	log.Info("Server exited gracefully")
	return nil
}
