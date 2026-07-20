package config

import (
	"errors"
	"strings"
)

func (c *Config) Validate() error {
	if err := c.validateApp(); err != nil {
		return err
	}
	if err := c.validateServer(); err != nil {
		return err
	}
	if err := c.validateDB(); err != nil {
		return err
	}
	if err := c.validateRedis(); err != nil {
		return err
	}
	if err := c.validateLog(); err != nil {
		return err
	}
	if err := c.validateJWT(); err != nil {
		return err
	}
	if err := c.validateProduction(); err != nil {
		return err
	}
	return nil
}

func (c *Config) validateApp() error {
	if c.App.Name == "" {
		return errors.New("app.name is required")
	}
	if c.App.Version == "" {
		return errors.New("app.version is required")
	}
	switch c.App.Environment {
	case "development", "staging", "production":
	default:
		return errors.New("app.environment must be one of: development, staging, production")
	}
	return nil
}

func (c *Config) validateServer() error {
	if c.Server.Port == "" {
		return errors.New("server.port is required")
	}
	if c.Server.ReadTimeout < 0 {
		return errors.New("server.readtimeout must be non-negative")
	}
	if c.Server.WriteTimeout < 0 {
		return errors.New("server.writetimeout must be non-negative")
	}
	if c.Server.IdleTimeout < 0 {
		return errors.New("server.idletimeout must be non-negative")
	}
	if c.Server.ShutdownTimeout < 0 {
		return errors.New("server.shutdowntimeout must be non-negative")
	}
	if c.Server.MaxHeaderBytes < 0 {
		return errors.New("server.maxheaderbytes must be non-negative")
	}
	return nil
}

func (c *Config) validateDB() error {
	if c.DB.Host == "" {
		return errors.New("db.host is required")
	}
	if c.DB.ContainerHost == "" {
		return errors.New("db.container_host is required")
	}
	if c.DB.Port <= 0 || c.DB.Port > 65535 {
		return errors.New("db.port must be between 1 and 65535")
	}
	if c.DB.User == "" {
		return errors.New("db.user is required")
	}
	if c.DB.Name == "" {
		return errors.New("db.name is required")
	}
	if c.DB.MaxOpenConns <= 0 {
		return errors.New("db.max_open_conns must be greater than 0")
	}
	if c.DB.MaxIdleConns < 0 {
		return errors.New("db.max_idle_conns cannot be negative")
	}
	if c.DB.MaxIdleConns > c.DB.MaxOpenConns {
		return errors.New("db.max_idle_conns cannot be greater than db.max_open_conns")
	}
	if c.DB.ConnMaxLifetime < 0 {
		return errors.New("db.conn_max_lifetime cannot be negative")
	}
	if c.DB.ConnMaxIdleTime < 0 {
		return errors.New("db.conn_max_idle_time cannot be negative")
	}
	switch strings.ToLower(c.DB.SSLMode) {
	case "disable", "allow", "prefer", "require", "verify-ca", "verify-full":
	default:
		return errors.New("invalid db.ssl_mode")
	}
	return nil
}

func (c *Config) validateRedis() error {
	if c.Redis.Host == "" {
		return errors.New("redis.host is required")
	}
	if c.Redis.Port == "" {
		return errors.New("redis.port is required")
	}
	if c.Redis.DB < 0 {
		return errors.New("redis.db cannot be negative")
	}
	if c.Redis.DialTimeout <= 0 {
		return errors.New("redis.dial_timeout must be greater than 0")
	}
	if c.Redis.ReadTimeout <= 0 {
		return errors.New("redis.read_timeout must be greater than 0")
	}
	if c.Redis.WriteTimeout <= 0 {
		return errors.New("redis.write_timeout must be greater than 0")
	}
	return nil
}

func (c *Config) validateLog() error {
	switch strings.ToLower(c.Log.Format) {
	case "text", "json":
	default:
		return errors.New("log.format must be 'text' or 'json'")
	}
	return nil
}

func (c *Config) validateJWT() error {
	if c.JWT.AccessTokenTTL <= 0 {
		return errors.New("jwt access token ttl must be greater than 0")
	}
	if c.JWT.RefreshTokenTTL <= 0 {
		return errors.New("jwt refresh token ttl must be greater than 0")
	}
	return nil
}

func (c *Config) validateProduction() error {
	if c.App.Environment != "production" {
		return nil
	}
	if c.App.Debug {
		return errors.New("app.debug must be false in production")
	}
	if c.DB.Password == "" {
		return errors.New("db.password is required in production")
	}
	if c.DB.SSLMode == "disable" {
		return errors.New("db.ssl_mode cannot be 'disable' in production")
	}
	return nil
}
