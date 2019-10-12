package rule

import (
	"encoding/json"
	"fmt"
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/process"
	"time"

	"go.uber.org/zap"
)

const Pick = "pick"

type PickConfig struct {
	BetStep         uint          `json:"bet_step" mapstructure:"bet_step"`
	BasePrice       uint          `json:"base_price" validate:"required,gt=0"`
	ConfirmDuration time.Duration `mapstructure:"confirm_duration" validate:"gt=0"`
}

type pick struct {
	*base
	*PickConfig
}

func newPick(interval process.Interval, defaultConfig PickConfig) *pick {
	return &pick{newBase(Pick, interval), &defaultConfig}
}

func (rule *pick) UnmarshalJSON(data []byte) error {
	config := struct {
		*PickConfig
		ConfirmDuration core.Duration `json:"confirm_duration"`
	}{
		PickConfig:      rule.PickConfig,
		ConfirmDuration: core.Duration(rule.PickConfig.ConfirmDuration),
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("config extra duration unmarshal error: %s", err)
	}

	rule.PickConfig.ConfirmDuration = time.Duration(config.ConfirmDuration)

	return nil
}

func (rule pick) MarshalJSON() ([]byte, error) {
	config := struct {
		*PickConfig
		ConfirmDuration core.Duration `json:"confirm_duration"`
	}{
		PickConfig:      rule.PickConfig,
		ConfirmDuration: core.Duration(rule.PickConfig.ConfirmDuration),
	}

	data, err := json.Marshal(&config)
	if err != nil {
		return nil, fmt.Errorf("config extra marshal error: %s", err)
	}

	return data, nil
}

func (rule *pick) Config() interface{} {
	return rule.PickConfig
}

func (rule *pick) Sync(lot *core.Lot) {
	lot.BetStep = rule.BetStep
	lot.BasePrice = rule.BasePrice
	lot.CurrentPrice = rule.BasePrice
	lot.RulePrice = rule.BasePrice
	rule.base.Sync(lot)
}

func (rule *pick) Start(prx process.RuleProxy) error {
	logger := rule.logger.Named("start")

	lot := prx.ProcessLot()

	if lot.BookedAt != nil {
		lot.BookedAt = nil

		if err := prx.SaveLot(lot); err != nil {
			logger.Error("save lot failed", zap.Error(err))
			return err
		}

		if err := prx.ResetDateBook(); err != nil {
			logger.Error("reset date book failed", zap.Error(err))
			return err
		}
	}

	if err := prx.ClearBets(lot.ID); err != nil {
		logger.Error("clear bets failed", zap.Error(err))
		return err
	}

	lot.ClearBets()

	return nil
}

func (rule *pick) PlaceBet(act process.PlaceBet, prx process.RuleProxy) error {
	logger := rule.logger.Named("place_bet")

	lot := prx.ProcessLot()

	if act.Value() > rule.BasePrice {
		logger.Warn("bet invalid")
		return betInvalid
	}

	newBet := &core.Bet{
		Value:  act.Value(),
		Winner: true,
		LotID:  lot.ID,
		UserID: act.Executor().ID,
	}

	if err := prx.CreateBet(newBet); err != nil {
		logger.Error("create bet failed", zap.Error(err))
		return err
	}

	lot.AddBet(newBet)

	n := process.Now()

	lot.BookedAt = &n

	if err := prx.SaveLot(lot); err != nil {
		logger.Error("save lot failed", zap.Error(err))
		return err
	}

	if err := prx.SetDateBook(n, newBet); err != nil {
		logger.Error("set date book failed", zap.Error(err))
		return err
	}

	r, err := NewConfirm(rule.ConfirmDuration)
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
