package config

import (
	"github.com/azhai/gobus/environ"
	"github.com/azhai/goent/drivers"
)

type Config struct {
	Server   ServerConfig
	Database drivers.DatabaseConfig
	Auth     AuthConfig
	Admin    AdminConfig
}

type ServerConfig struct {
	Mode            string
	Port            int
	BaseURL         string
	SiteTitle       string
	SiteDescription string
}

type AuthConfig struct {
	JWTSecret   string
	TokenExpiry int // 小时
}

type AdminConfig struct {
	Username string
	Password string
	Email    string
}

var cfg *Config

func Load(env *environ.Environ) *Config {
	if cfg != nil {
		return cfg
	}
	if env == nil {
		env = environ.NewEmptyEnviron()
	}
	cfg = &Config{
		Server: ServerConfig{
			Mode:            env.GetStr("APP_MODE", "debug"),
			Port:            env.GetInt("SERVER_PORT", 8090),
			BaseURL:         env.GetStr("BASE_URL", "http://127.0.0.1:8090"),
			SiteTitle:       env.GetStr("SITE_TITLE", "Flock"),
			SiteDescription: env.GetStr("SITE_DESCRIPTION", "一个极简论坛"),
		},
		Database: drivers.LoadConfig(env, "flock.db"),
		Auth: AuthConfig{
			JWTSecret:   env.GetStr("JWT_SECRET", "change-me-in-production"),
			TokenExpiry: env.GetInt("TOKEN_EXPIRY", 72),
		},
		Admin: AdminConfig{
			Username: env.GetStr("ADMIN_USERNAME", "admin"),
			Password: env.GetStr("ADMIN_PASSWORD", ""),
			Email:    env.GetStr("ADMIN_EMAIL", "admin@flock.local"),
		},
	}
	return cfg
}

func GetConfig() *Config {
	return cfg
}

func GetJWTSecret() string         { return cfg.Auth.JWTSecret }
func GetTokenExpiry() int          { return cfg.Auth.TokenExpiry }
func GetSiteTitle() string         { return cfg.Server.SiteTitle }
func GetServerInfo() (int, string) { return cfg.Server.Port, cfg.Server.BaseURL }
