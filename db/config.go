package db

import (
	"errors"
	"strings"
)

const (
	DriverMSSQL  = "mssql"
	DriverMySQL  = "mysql"
	DriverSQLite = "sqlite3"
)

type Config struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

func (cfg *Config) Validate() error {
	if cfg.Driver != DriverSQLite && len(strings.TrimSpace(cfg.Username)) == 0 {
		return errors.New("username not set")
	}
	if cfg.Driver != DriverSQLite && len(strings.TrimSpace(cfg.Password)) == 0 {
		return errors.New("password not set")
	}
	if len(strings.TrimSpace(cfg.Database)) == 0 {
		return errors.New("database not set")
	}
	return nil
}

func DefaultConfig() *Config {
	return &Config{
		Driver:   DriverSQLite,
		Database: "../../database/database.sqlite",
	}
}
