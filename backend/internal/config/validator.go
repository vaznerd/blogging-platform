package config

import (
	"fmt"
	"strings"
)

func (c *Config) Validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}

	if c.App.Version == "" {
		return fmt.Errorf("app.version is required")
	}

	switch c.App.Environment {
	case "development", "staging", "production":
	default:
		return fmt.Errorf("app.environment must be one of: development, staging, production")
	}

	if c.Server.Port == "" {
		return fmt.Errorf("server.port is required")
	}

	if c.Server.ReadTimeout < 0 {
		return fmt.Errorf("server.readtimeout must be non-negative")
	}

	if c.Server.WriteTimeout < 0 {
		return fmt.Errorf("server.writetimeout must be non-negative")
	}

	if c.Server.IdleTimeout < 0 {
		return fmt.Errorf("server.idletimeout must be non-negative")
	}

	if c.Server.ShutdownTimeout < 0 {
		return fmt.Errorf("server.shutdowntimeout must be non-negative")
	}

	if c.Server.MaxHeaderBytes < 0 {
		return fmt.Errorf("server.maxheaderbytes must be non-negative")
	}

	if c.DB.Host == "" {
		return fmt.Errorf("db.host is required")
	}

	if c.DB.HostGo == "" {
		return fmt.Errorf("db.hostGo is required")
	}

	if c.DB.Port <= 0 || c.DB.Port > 65535 {
		return fmt.Errorf("db.port must be between 1 and 65535")
	}

	if c.DB.User == "" {
		return fmt.Errorf("db.user is required")
	}

	if c.DB.Name == "" {
		return fmt.Errorf("db.name is required")
	}

	if c.DB.MaxOpenConns <= 0 {
		return fmt.Errorf("db.max_open_conns must be greater than 0")
	}

	if c.DB.MaxIdleConns < 0 {
		return fmt.Errorf("db.max_idle_conns cannot be negative")
	}

	if c.DB.MaxIdleConns > c.DB.MaxOpenConns {
		return fmt.Errorf("db.max_idle_conns cannot be greater than db.max_open_conns")
	}

	if c.DB.ConnMaxLifetime < 0 {
		return fmt.Errorf("db.conn_max_lifetime cannot be negative")
	}

	if c.DB.ConnMaxIdleTime < 0 {
		return fmt.Errorf("db.conn_max_idle_time cannot be negative")
	}

	switch strings.ToLower(c.DB.SSLMode) {
	case "disable", "allow", "prefer", "require", "verify-ca", "verify-full":
	default:
		return fmt.Errorf("invalid db.ssl_mode")
	}

	if c.Redis.Host == "" {
		return fmt.Errorf("redis.host is required")
	}

	if c.Redis.Port == "" {
		return fmt.Errorf("redis.port is required")
	}

	if c.Redis.DB < 0 {
		return fmt.Errorf("redis.db cannot be negative")
	}

	if c.Redis.DialTimeout <= 0 {
		return fmt.Errorf("redis.dial_timeout must be greater than 0")
	}

	if c.Redis.ReadTimeout <= 0 {
		return fmt.Errorf("redis.read_timeout must be greater than 0")
	}

	if c.Redis.WriteTimeout <= 0 {
		return fmt.Errorf("redis.write_timeout must be greater than 0")
	}

	switch strings.ToLower(c.Log.Format) {
	case "text", "json":
	default:
		return fmt.Errorf("log.format must be 'text' or 'json'")
	}

	if c.App.Environment == "production" {
		if c.App.Debug {
			return fmt.Errorf("app.debug must be false in production")
		}
		if c.DB.Password == "" {
			return fmt.Errorf("db.password is required in production")
		}
		if c.DB.SSLMode == "disable" {
			return fmt.Errorf("db.ssl_mode cannot be 'disable' in production")
		}
	}

	return nil
}
