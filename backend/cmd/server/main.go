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
	"time"

	"codeberg.org/vaznerd/blogging-platform/internal/auth"
	"codeberg.org/vaznerd/blogging-platform/internal/category"
	"codeberg.org/vaznerd/blogging-platform/internal/config"
	"codeberg.org/vaznerd/blogging-platform/internal/logger"
	"codeberg.org/vaznerd/blogging-platform/internal/server"
	"codeberg.org/vaznerd/blogging-platform/internal/tag"
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

func bootstrap() (*slog.Logger, *config.Config, *resend.Client, error) {
	bootstrapLogger := slog.Default()
	bootstrapLogger.Info("Starting Go Blogging platform API...")
	cfg, err := config.LoadConfig()
	if err != nil {
		bootstrapLogger.Error("Failed to load configuration", "error", err)
		return nil, nil, nil, err
	}
	if valErr := cfg.Validate(); valErr != nil {
		bootstrapLogger.Error("Configuration validation failed", "error", valErr)
		return nil, nil, nil, valErr
	}
	log := logger.NewLogger(&cfg.Log)
	cfg.LogAllConfig(log)

	mail := resend.NewClient(cfg.Resend.APIKey)

	return log, cfg, mail, nil
}

func connectDB(cfg *config.Config, log *slog.Logger) (*pgxpool.Pool, error) {
	q := url.Values{}
	q.Set("sslmode", cfg.DB.SSLMode)
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.DB.User, cfg.DB.Password),
		Host:     net.JoinHostPort(cfg.DB.Host, strconv.Itoa(cfg.DB.Port)),
		Path:     cfg.DB.Name,
		RawQuery: q.Encode(),
	}
	dbconfig, err := pgxpool.ParseConfig(u.String())
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}
	dbconfig.MaxConns = cfg.DB.MaxOpenConns
	dbconfig.MinConns = 5
	dbconfig.MaxConnLifetime = cfg.DB.ConnMaxLifetime
	dbconfig.MaxConnIdleTime = cfg.DB.ConnMaxIdleTime

	dbpool, err := pgxpool.NewWithConfig(context.Background(), dbconfig)
	if err != nil {
		return nil, fmt.Errorf("connect to db: %w", err)
	}
	if pingErr := dbpool.Ping(context.Background()); pingErr != nil {
		dbpool.Close()
		return nil, fmt.Errorf("ping postgres: %w", pingErr)
	}
	log.Info("postgress connection established")

	return dbpool, nil
}

func connectRedis(cfg *config.Config, log *slog.Logger) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if pingErr := rdb.Ping(context.Background()).Err(); pingErr != nil {
		if closeErr := rdb.Close(); closeErr != nil {
			log.Error("rdb.Close", "error", closeErr)
		}
		return nil, fmt.Errorf("ping redis: %w", pingErr)
	}
	log.Info("redis connection established")

	return rdb, nil
}

func startSessionCleanup(authService *auth.Service, log *slog.Logger) {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			rows, err := authService.DeleteExpiredSessions(context.Background())
			if err != nil {
				log.Error("failed to cleanup expired sessions", "error", err)
			} else if rows > 0 {
				log.Info("cleaned up expired sessions", "rows_deleted", rows)
			}
		}
	}()
}

func run() error {
	log, cfg, mail, err := bootstrap()
	if err != nil {
		return err
	}

	dbpool, err := connectDB(cfg, log)
	if err != nil {
		return err
	}

	rdb, err := connectRedis(cfg, log)
	if err != nil {
		dbpool.Close()
		return err
	}

	userRepository := user.NewRepository(dbpool)
	userService := user.NewService(userRepository)

	authRepository := auth.NewRefreshTokenRepository(dbpool)
	authService := auth.NewService(&cfg.JWT, authRepository, userService)

	categoryRepository := category.NewRepository(dbpool)
	categoryService := category.NewService(categoryRepository)

	tagRepository := tag.NewRepository(dbpool)
	tagService := tag.NewService(tagRepository)

	router := server.NewRouter(
		userService,
		authService,
		categoryService,
		tagService,
		log,
		mail,
		cfg.App.Debug,
		cfg.App.FrontendURL,
		cfg.App.FrontendURL,
		cfg.Server.TrustedProxies,
	)
	srv := &http.Server{
		Addr:           ":" + cfg.Server.Port,
		Handler:        router,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		IdleTimeout:    cfg.Server.IdleTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	go func() {
		log.Info("Server starting", "address", srv.Addr)
		log.Info("Health check available",
			"url",
			fmt.Sprintf("http://localhost:%s/health", cfg.Server.Port))
		if srvErr := srv.ListenAndServe(); srvErr != nil && srvErr != http.ErrServerClosed {
			log.Error("Server error", "error", srvErr)
			os.Exit(1)
		}
	}()

	startSessionCleanup(authService, log)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	log.Info("Received shutdown signal", "signal", sig)
	log.Info("Shutting down server gracefully...")

	if shutdownErr := srv.Shutdown(ctx); shutdownErr != nil {
		log.Error("Server forced to shutdown", "error", shutdownErr)
		return shutdownErr
	}
	dbpool.Close()
	if closeErr := rdb.Close(); closeErr != nil {
		log.Error("rdb.Close", "error", closeErr)
	}

	log.Info("Server exited gracefully")
	return nil
}
