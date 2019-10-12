package rule

import (
	"encoding/json"
	"fmt"
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/process"
	"time"

	"go.uber.org/zap"
)

const Normal = "normal"

type NormalConfig struct {
	BetStep         uint          `json:"bet_step" mapstructure:"bet_step"`
	BasePrice       uint          `json:"base_price" validate:"required,gt=0"`
	LastMoment      time.Duration `mapstructure:"last_moment" validate:"gte=0"`
	ProlongDuration time.Duration `mapstructure:"prolong_duration" validate:"gt=0"`
	MaxDuration     time.Duration `mapstructure:"max_duration" validate:"gt=0"`
	ConfirmDuration time.Duration `mapstructure:"confirm_duration" validate:"gt=0"`
}

func (conf NormalConfig) basePrice() uint {
	return conf.BasePrice
}

func (conf NormalConfig) price() uint {
	return conf.BasePrice
}

func (conf NormalConfig) betStep() uint {
	return conf.BetStep
}

func (conf NormalConfig) confirmDuration() time.Duration {
	return conf.ConfirmDuration
}

func (conf NormalConfig) extra() bool {
	return false
}

type normal struct {
	*simple
	*NormalConfig
}

func newNormal(interval process.Interval, defaultConfig NormalConfig) *normal {
	return &normal{newSimple(Normal, interval, &defaultConfig), &defaultConfig}
}

func (rule *normal) UnmarshalJSON(data []byte) error {
	config := struct {
		*NormalConfig
		LastMoment      core.Duration `json:"last_moment"`
		ProlongDuration core.Duration `json:"prolong_duration"`
		MaxDuration     core.Duration `json:"max_duration"`
		ConfirmDuration core.Duration `json:"confirm_duration"`
	}{
		NormalConfig:    rule.NormalConfig,
		LastMoment:      core.Duration(rule.NormalConfig.LastMoment),
		ProlongDuration: core.Duration(rule.NormalConfig.ProlongDuration),
		MaxDuration:     core.Duration(rule.NormalConfig.MaxDuration),
		ConfirmDuration: core.Duration(rule.NormalConfig.ConfirmDuration),
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("config extra duration unmarshal error: %s", err)
	}

	rule.NormalConfig.LastMoment = time.Duration(config.LastMoment)
	rule.NormalConfig.ProlongDuration = time.Duration(config.ProlongDuration)
	rule.NormalConfig.MaxDuration = time.Duration(config.MaxDuration)
	rule.NormalConfig.ConfirmDuration = time.Duration(config.ConfirmDuration)

	return nil
}

func (rule normal) MarshalJSON() ([]byte, error) {
	config := struct {
		*NormalConfig
		LastMoment      core.Duration `json:"last_moment"`
		ProlongDuration core.Duration `json:"prolong_duration"`
		MaxDuration     core.Duration `json:"max_duration"`
		ConfirmDuration core.Duration `json:"confirm_duration"`
	}{
		NormalConfig:    rule.NormalConfig,
		LastMoment:      core.Duration(rule.NormalConfig.LastMoment),
		ProlongDuration: core.Duration(rule.NormalConfig.ProlongDuration),
		MaxDuration:     core.Duration(rule.NormalConfig.MaxDuration),
		ConfirmDuration: core.Duration(rule.NormalConfig.ConfirmDuration),
	}

	data, err := json.Marshal(&config)
	if err != nil {
		return nil, fmt.Errorf("config extra marshal error: %s", err)
	}

	return data, nil
}

func (rule *normal) Config() interface{} {
	return rule.NormalConfig
}

func (rule *normal) Sync(lot *core.Lot) {
	lot.CurrentPrice = rule.price()
	rule.simple.Sync(lot)
}

func (rule *normal) PlaceBet(act process.PlaceBet, prx process.RuleProxy) error {
	logger := rule.logger.Named("place_bet")

	if err := rule.simple.PlaceBet(act, prx); err != nil {
		return err
	}

	rest := prx.End().Sub(now())
	lastMoment := rule.LastMoment
	maxTime := rule.interval.Start().Add(rule.MaxDuration)
	restMaxTime := maxTime.Sub(now())

	if rest < lastMoment && restMaxTime > lastMoment {
		d := rule.ProlongDuration

		prx.Prolong(d)

		logger.Debug("prolong", zap.Duration("duration", d))
	}

	return nil
}
