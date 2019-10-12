package rule

import (
	"time"
)

const DefaultConfirmDuration = 1 * time.Hour

type Config struct {
	Normal *NormalConfig `mapstructure:"normal"`
	Extra  *ExtraConfig  `mapstructure:"extra"`
	Pick   *PickConfig   `mapstructure:"pick"`
}

func DefaultConfig() *Config {
	return &Config{
		Normal: &NormalConfig{
			BetStep:         500,
			LastMoment:      15 * time.Minute,
			ProlongDuration: 15 * time.Minute,
			MaxDuration:     3 * time.Hour,
			ConfirmDuration: DefaultConfirmDuration,
		},
		Extra: &ExtraConfig{
			BetStep:         500,
			HotCount:        3,
			HotDuration:     30 * time.Minute,
			ConfirmDuration: DefaultConfirmDuration,
		},
		Pick: &PickConfig{
			BetStep:         500,
			ConfirmDuration: DefaultConfirmDuration,
		},
	}
}
