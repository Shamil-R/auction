package auction

import "gitlab/nefco/auction/process/rule"

type Config struct {
	RuleConfig *rule.Config `mapstructure:"rule"`
}

func DefaultConfig() *Config {
	return &Config{
		RuleConfig: rule.DefaultConfig(),
	}
}
