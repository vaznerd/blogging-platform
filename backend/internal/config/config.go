package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	App    AppConfig    `koanf:"app"`
	Server ServerConfig `koanf:"server"`
	Log    LogConfig    `koanf:"log"`
	DB     DBConfig     `koanf:"db"`
	Redis  RedisConfig  `koanf:"redis"`
	Resend ResendConfig `koanf:"resend"`
	JWT    JWTConfig    `koanf:"jwt"`
}

type AppConfig struct {
	Name        string `koanf:"name"`
	Version     string `koanf:"version"`
	Environment string `koanf:"environment"`
	Debug       bool   `koanf:"debug"`
}

type ServerConfig struct {
	Port            string        `koanf:"port"`
	ReadTimeout     time.Duration `koanf:"read_timeout"`
	WriteTimeout    time.Duration `koanf:"write_timeout"`
	IdleTimeout     time.Duration `koanf:"idle_timeout"`
	ShutdownTimeout time.Duration `koanf:"shutdown_timeout"`
	MaxHeaderBytes  int           `koanf:"max_header_bytes"`
}

type LogConfig struct {
	Level     slog.Level `koanf:"level"`
	Format    string     `koanf:"format"`
	AddSource bool       `koanf:"add_source"`
}

type DBConfig struct {
	Host            string        `koanf:"host"`
	HostGo          string        `koanf:"host_go"`
	Port            int           `koanf:"port"`
	User            string        `koanf:"user"`
	Password        string        `koanf:"password"`
	Name            string        `koanf:"name"`
	MaxOpenConns    int32         `koanf:"max_open_conns"`
	MaxIdleConns    int32         `koanf:"max_idle_conns"`
	ConnMaxLifetime time.Duration `koanf:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `koanf:"conn_max_idle_time"`
	SSLMode         string        `koanf:"ssl_mode"`
}

type RedisConfig struct {
	DB           int           `koanf:"db"`
	Host         string        `koanf:"host"`
	HostGo       string        `koanf:"host_go"`
	Port         string        `koanf:"port"`
	Password     string        `koanf:"password"`
	DialTimeout  time.Duration `koanf:"dial_timeout"`
	ReadTimeout  time.Duration `koanf:"read_timeout"`
	WriteTimeout time.Duration `koanf:"write_timeout"`
}

type JWTConfig struct {
	Secret          string        `koanf:"jwt_secret"`
	AccessTokenTTL  time.Duration `koanf:"access_token_ttl"`
	RefreshTokenTTL time.Duration `koanf:"refresh_token_ttl"`
}

type ResendConfig struct {
	APIKey string `koanf:"api_key"`
}

func LoadConfig() (*Config, error) {
	k := koanf.New(".")
	cfg := &Config{}

	if err := k.Load(file.Provider("configs/config.yaml"), yaml.Parser()); err != nil {
		return nil, err
	}
	k.Load(env.Provider("",
		".",
		func(s string) string {
			return strings.ToLower(strings.ReplaceAll(s, "__", "."))
		},
	), nil)

	if err := k.Unmarshal("", cfg); err != nil {
		return nil, err
	}

	if val := os.Getenv("RESEND_API"); val != "" {
		cfg.Resend.APIKey = val
	}
	if cfg.Resend.APIKey == "" {
		return cfg, fmt.Errorf("resend api key not found")
	}
	if val := os.Getenv("JWT_SECRET"); val != "" {
		cfg.JWT.Secret = val
	}
	if cfg.JWT.Secret == "" {
		return cfg, fmt.Errorf("jwt secret not found")
	}

	return cfg, nil
}

func (c *Config) LogAllConfig(log *slog.Logger) {
	log.Info("Loaded Configuration:")
	log.Info("App", "Name", c.App.Name, "Version", c.App.Version, "Environment", c.App.Environment, "Debug", c.App.Debug)
	log.Info("Server", "Port", c.Server.Port, "ReadTimeout", c.Server.ReadTimeout, "WriteTimeout", c.Server.WriteTimeout, "IdleTimeout", c.Server.IdleTimeout, "ShutdownTimeout", c.Server.ShutdownTimeout, "MaxHeaderBytes", c.Server.MaxHeaderBytes)
	log.Info("Log", "Level", c.Log.Level, "Format", c.Log.Format, "AddSource", c.Log.AddSource)
	log.Info("Database", "Host", c.DB.Host, "HostGo", c.DB.HostGo, "Port", c.DB.Port, "User", c.DB.User, "Password", "<redacted>", "Name", c.DB.Name, "MaxOpenConns", c.DB.MaxOpenConns, "MaxIdleConns", c.DB.MaxIdleConns, "ConnMaxLifetime", c.DB.ConnMaxLifetime, "ConnMaxIdleTime", c.DB.ConnMaxIdleTime, "SSLMode", c.DB.SSLMode)
	log.Info("Redis", "DB", c.Redis.DB, "Host", c.Redis.Host, "HostGo", c.Redis.HostGo, "Port", c.Redis.Port, "Password", "<redacted>", "DialTimeout", c.Redis.DialTimeout, "ReadTimeout", c.Redis.ReadTimeout, "WriteTimeout", c.Redis.WriteTimeout)
	log.Info("Resend", "APIKey", "<redacted>")
	log.Info("JWT", "Secret", "<redacted>", "AccessTokenTTL", c.JWT.AccessTokenTTL, "RefreshTokenTTL", c.JWT.RefreshTokenTTL)
}
