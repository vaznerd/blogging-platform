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

type ResendConfig struct {
	APIKey string `koanf:"api_key"`
}

func LoadConfig() (*Config, error) {
	k := koanf.New(".")
	cfg := &Config{}

	if err := k.Load(file.Provider("config/config.yaml"), yaml.Parser()); err != nil {
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

	return cfg, nil
}
