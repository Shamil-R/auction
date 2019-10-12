package server

import (
	"time"
)

type Config struct {
	Port              uint          `mapstructure:"port"`
	JWTSecret         string        `mapstructure:"jwt_secret"`
	ManagerJWTExpires time.Duration `mapstructure:"manager_jwt_expires"`
	UserJWTExpires    time.Duration `mapstructure:"user_jwt_expires"`
}

func DefaultConfig() *Config {
	return &Config{
		Port:              8080,
		JWTSecret:         "secret",
		ManagerJWTExpires: 24 * 365 * 10 * time.Hour,
		UserJWTExpires:    24 * time.Hour,
	}
}
