package rule

import (
	"encoding/json"
	"fmt"
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/process"
	"time"

	"go.uber.org/zap"
)

const Extra = "extra"

type ExtraConfig struct {
	BetStep         uint          `json:"bet_step" mapstructure:"bet_step"`
	BasePrice       uint          `json:"base_price" validate:"required,gt=0"`
	ExtraPrice      uint          `json:"extra_price" db:"extra_price" validate:"required,gt=0"`
	HotCount        uint          `json:"hot_count" mapstructure:"hot_count" validate:"gt=0"`
	HotDuration     time.Duration `mapstructure:"hot_duration" validate:"gt=0"`
	ConfirmDuration time.Duration `mapstructure:"confirm_duration" validate:"gt=0"`
}

func (conf ExtraConfig) basePrice() uint {
	return conf.BasePrice
}

func (conf ExtraConfig) price() uint {
	return conf.ExtraPrice
}

func (conf ExtraConfig) betStep() uint {
	return conf.BetStep
}

func (conf ExtraConfig) confirmDuration() time.Duration {
	return conf.ConfirmDuration
}

func (conf ExtraConfig) extra() bool {
	return true
}

type extra struct {
	*simple
	*ExtraConfig
}

func newExtra(interval process.Interval, defaultConfig ExtraConfig) *extra {
	return &extra{
		simple:      newSimple(Extra, interval, &defaultConfig),
		ExtraConfig: &defaultConfig,
	}
}

func (rule *extra) UnmarshalJSON(data []byte) error {
	config := struct {
		*ExtraConfig
		HotDuration     core.Duration `json:"hot_duration"`
		ConfirmDuration core.Duration `json:"confirm_duration"`
	}{
		ExtraConfig:     rule.ExtraConfig,
		HotDuration:     core.Duration(rule.ExtraConfig.HotDuration),
		ConfirmDuration: core.Duration(rule.ExtraConfig.ConfirmDuration),
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("config extra duration unmarshal error: %s", err)
	}

	rule.ExtraConfig.HotDuration = time.Duration(config.HotDuration)
	rule.ExtraConfig.ConfirmDuration = time.Duration(config.ConfirmDuration)

	return nil
}

func (rule extra) MarshalJSON() ([]byte, error) {
	config := struct {
		*ExtraConfig
		HotDuration     core.Duration `json:"hot_duration"`
		ConfirmDuration core.Duration `json:"confirm_duration"`
	}{
		ExtraConfig:     rule.ExtraConfig,
		HotDuration:     core.Duration(rule.ExtraConfig.HotDuration),
		ConfirmDuration: core.Duration(rule.ExtraConfig.ConfirmDuration),
	}

	data, err := json.Marshal(&config)
	if err != nil {
		return nil, fmt.Errorf("config extra marshal error: %s", err)
	}

	return data, nil
}

func (rule *extra) Config() interface{} {
	return rule.ExtraConfig
}

func (rule *extra) Stop(prx process.RuleProxy) error {
	return rule.base.Stop(prx)
}

func (rule *extra) PlaceBet(act process.PlaceBet, prx process.RuleProxy) error {
	logger := rule.logger.Named("place_bet")

	lot := prx.ProcessLot()

	if err := rule.simple.PlaceBet(act, prx); err != nil {
		return err
	}

	unique := make(map[uint]bool)

	for _, bet := range lot.Bets {
		if _, ok := unique[bet.UserID]; !ok {
			unique[bet.UserID] = true
		}
	}

	if uint(len(unique)) >= rule.HotCount {
		rule, err := newHot(rule.ExtraConfig)
		if err != nil {
			logger.Error("hot rule failed", zap.Error(err))
			return err
		}
		if err := prx.Run(rule); err != nil {
			logger.Error("run rule failed", zap.Error(err))
			return err
		}
		logger.Debug("run hot rule")
		return nil
	}

	curBet := lot.CurrentBet()
	if curBet != nil && curBet.Value <= rule.basePrice() {
		curBet.Winner = true

		if err := prx.SaveBet(curBet); err != nil {
			logger.Error("save bet failed", zap.Error(err))
			return err
		}

		n := process.Now()

		lot.BookedAt = &n

		if err := prx.SaveLot(lot); err != nil {
			logger.Error("save lot failed", zap.Error(err))
			return err
		}

		if err := prx.SetDateBook(n, curBet); err != nil {
			logger.Error("set date book failed", zap.Error(err))
			return err
		}

		r, err := NewConfirm(rule.confirmDuration())
		if err != nil {
			logger.Error("confirm rule failed", zap.Error(err))
			return err
		}

		if err := prx.Run(r); err != nil {
			logger.Error("run rule failed", zap.Error(err))
			return err
		}
	}

	return nil
}

func (rule *extra) AcceptBet(act process.AcceptBet, prx process.RuleProxy) error {
	logger := rule.logger.Named("accept_bet")

	lot := act.GetLot()

	bet := lot.BetByID(act.BetID())
	if bet == nil {
		logger.Warn("bet not found")
		return betNotFound
	}

	bet.Winner = true

	if err := prx.SaveBet(bet); err != nil {
		logger.Error("save bet failed", zap.Error(err))
		return err
	}

	n := process.Now()

	lot.BookedAt = &n

	if err := prx.SaveLot(lot); err != nil {
		logger.Error("save lot failed", zap.Error(err))
		return err
	}

	if err := prx.SetDateBook(n, bet); err != nil {
		logger.Error("set date book failed", zap.Error(err))
		return err
	}

	r, err := NewConfirm(rule.confirmDuration())
	if err != nil {
		logger.Error("confirm rule failed", zap.Error(err))
		return err
	}

	if err := prx.Run(r); err != nil {
		logger.Error("run rule failed", zap.Error(err))
		return err
	}

	return nil
}
