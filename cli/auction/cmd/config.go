package cmd

import (
	"errors"
	"gitlab/nefco/auction/auction"
	"gitlab/nefco/auction/db"
	"gitlab/nefco/auction/server"
	"gitlab/nefco/auction/service"

	"github.com/fatih/structs"
)

type config struct {
	Production            bool                           `mapstructure:"production"`
	LogLevel              string                         `mapstructure:"log_level"`
	RootPassword          string                         `mapstructure:"root_password"`
	ServerConfig          *server.Config                 `mapstructure:"server"`
	DBConfig              *db.Config                     `mapstructure:"db"`
	AuctionConfig         *auction.Config                `mapstructure:"auction"`
	FeedbackServiceConfig *service.ConfigFeedbackService `mapstructure:"feedback_service"`
	BackServiceConfig     *service.ConfigBackService     `mapstructure:"back_service"`
}

func (cfg *config) validate() error {
	if err := cfg.DBConfig.Validate(); err != nil {
		return errors.New("db config: " + err.Error())
	}
	return nil
}

func defaultConfig() *config {
	return &config{
		Production:            false,
		LogLevel:              "DEBUG",
		RootPassword:          "password",
		ServerConfig:          server.DefaultConfig(),
		DBConfig:              db.DefaultConfig(),
		AuctionConfig:         auction.DefaultConfig(),
		FeedbackServiceConfig: service.DefaultConfigFeedbackService(),
		BackServiceConfig:     service.DefaultConfigBackService(),
	}
}

func configVars(s interface{}) []string {
	res := make([]string, 0, 1)

	fields := structs.Fields(s)

	for _, field := range fields {
		tag := field.Tag("mapstructure")

		if structs.IsStruct(field.Value()) {
			arr := configVars(field.Value())

			for _, t := range arr {
				res = append(res, tag+"."+t)
			}
		} else {
			res = append(res, tag)
		}
	}

	return res
}
