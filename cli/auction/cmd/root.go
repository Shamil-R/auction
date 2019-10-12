package cmd

import (
	"fmt"
	"gitlab/nefco/auction/auction"
	"gitlab/nefco/auction/db"
	"gitlab/nefco/auction/server"
	"gitlab/nefco/auction/service"
	"os"
	"strings"

	"go.uber.org/zap"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	cfgName = "auction"
)

var (
	cfgFile string
	cfg     = defaultConfig()
)

var RootCmd = &cobra.Command{
	Use:   "auction",
	Short: "auction server",
	RunE:  run,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		fmt.Sprintf("config file (default is $HOME/.%s.yaml)", cfgName),
	)
}

func initConfig() {
	if len(strings.TrimSpace(cfgFile)) != 0 {
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName("." + cfgName)
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath("./")

	viper.SetEnvPrefix(cfgName)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	for _, v := range configVars(cfg) {
		viper.BindEnv(v)
	}

	if err := viper.ReadInConfig(); err == nil {
	}

	prefix := "AUCTION_BACK_SERVICE_MANAGERS_"
	for _, pair := range os.Environ() {
		list := strings.Split(pair, "=")
		if len(list) == 2 {
			key := list[0]
			value := list[1]
			if strings.HasPrefix(key, prefix) {
				key = strings.TrimLeft(key, prefix)
				key = strings.ToLower(key)
				cfg.BackServiceConfig.Managers[key] = value
			}
		}
	}
}

func run(cmd *cobra.Command, args []string) error {
	if err := viper.Unmarshal(&cfg); err != nil {
		return err
	}

	if err := cfg.validate(); err != nil {
		return err
	}

	var logger *zap.Logger
	var err error
	if cfg.Production {
		c := zap.NewProductionConfig()
		c.Level.UnmarshalText([]byte(cfg.LogLevel))
		logger, err = c.Build()
	} else {
		c := zap.NewDevelopmentConfig()
		c.Level.UnmarshalText([]byte(cfg.LogLevel))
		logger, err = c.Build()
	}
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(logger)

	zap.L().Info(
		fmt.Sprintf("using config file: %s\n\n", viper.ConfigFileUsed()),
	)

	db, err := db.New(cfg.DBConfig)
	if err != nil {
		zap.L().Error("error", zap.Error(err))
		return err
	}

	user, err := service.CheckRootUser(cfg.RootPassword, db)
	if err != nil {
		zap.L().Error("error", zap.Error(err))
		return err
	}

	err = service.CheckManagers(cfg.BackServiceConfig.Managers, db)
	if err != nil {
		zap.L().Error("error", zap.Error(err))
		return err
	}

	// if !cfg.Production {
	// 	if err := service.Seed(db); err != nil {
	// 		zap.L().Error("error", zap.Error(err))
	// 		return err
	// 	}
	// }

	notify := service.NewNotifyService()

	feedbackSvc := service.NewFeedbackService(cfg.FeedbackServiceConfig)

	auction := auction.New(cfg.AuctionConfig, db, notify, feedbackSvc,
		cfg.BackServiceConfig)

	if err := auction.Restore(user); err != nil {
		zap.L().Error("error", zap.Error(err))
		return err
	}

	if err := server.NewServer(cfg.ServerConfig, auction, notify); err != nil {
		zap.L().Error("error", zap.Error(err))
		return err
	}

	return nil
}
