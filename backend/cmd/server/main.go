package server

import (
	"context"
	"log/slog"
	"net"
	"net/url"
	"os"
	"strconv"

	"codeberg.org/vaznerd/blogging-platform/internal/config"
	"codeberg.org/vaznerd/blogging-platform/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
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
	log := logger.NewLogger(cfg.Log)
	// write a method to the cfg in config.go to log all config

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
	}
	log.Info("redis connection established")

	return nil
}
