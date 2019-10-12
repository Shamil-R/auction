package rule

import (
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/process"
	"time"

	"go.uber.org/zap"
)

type commonConfig interface {
	basePrice() uint
	price() uint
	betStep() uint
	confirmDuration() time.Duration
	extra() bool
}

type common struct {
	*base
	commonConfig
}

func newCommon(name string, interval process.Interval, config commonConfig) *common {
	return &common{newBase(name, interval), config}
}

func (rule *common) PlaceBet(act process.PlaceBet, prx process.RuleProxy) error {
	logger := rule.logger.Named("place_bet")

	lot := prx.ProcessLot()

	curBet := lot.CurrentBet()

	var curPrice uint
	if curBet != nil {
		if curBet.UserID != act.Executor().ID && curBet.Value <= act.Value() {
			logger.Warn("invalid bid step")
			return betStepInvalid
		}
		curPrice = curBet.Value
	} else {
		if rule.extra() {
			curPrice = act.Value()
		} else {
			curPrice = rule.price()
		}
	}

	if ((curPrice-act.Value())%rule.betStep()) != 0 &&
		act.Value() != rule.basePrice() {
		logger.Warn("invalid bid step")
		return betStepInvalid
	}

	newBet := &core.Bet{
		Value:  act.Value(),
		LotID:  lot.ID,
		UserID: act.Executor().ID,
	}

	if err := prx.CreateBet(newBet); err != nil {
		logger.Error("create bet failed", zap.Error(err))
		return err
	}

	lot.AddBet(newBet)

	return nil
}

func (rule *common) CancelBet(act process.Action, prx process.RuleProxy) error {
	logger := rule.logger.Named("cancel_bet")

	lot := prx.ProcessLot()

	curBet := lot.CurrentBet()

	if curBet == nil {
		logger.Warn("bet not found")
		return betNotFound
	}

	if curBet.UserID != act.Executor().ID {
		logger.Warn("bet other user")
		return betOtherUser
	}

	if err := prx.DeleteBet(curBet); err != nil {
		logger.Error("delete bet failed", zap.Error(err))
		return err
	}

	lot.RemoveBet(curBet)

	return nil
}

func (rule *common) Sync(lot *core.Lot) {
	lot.BasePrice = rule.basePrice()
	lot.BetStep = rule.betStep()
	lot.RulePrice = rule.price()
	rule.base.Sync(lot)
}

func (rule *common) Stop(prx process.RuleProxy) error {
	logger := rule.logger.Named("stop")

	lot := prx.ProcessLot()

	curBet := lot.CurrentBet()

	if curBet != nil {
		if curBet.Value <= rule.price() {
			curBet.Winner = true

			if err := prx.SaveBet(curBet); err != nil {
				logger.Error("save bet failed", zap.Error(err))
				return err
			}

			n := now()

			lot.BookedAt = &n

			if err := prx.SaveLot(lot); err != nil {
				logger.Error("save lot failed", zap.Error(err))
				return err
			}

			if err := prx.SetDateBook(n, curBet); err != nil {
				logger.Error("set date book failed", zap.Error(err))
				return err
			}

			rule, err := NewConfirm(rule.confirmDuration())
			if err != nil {
				logger.Error("confirm rule failed", zap.Error(err))
				return err
			}

			if err := prx.Run(rule); err != nil {
				logger.Error("run rule failed", zap.Error(err))
				return err
			}

			return nil
		}

		if err := prx.LotNotWinner(lot.UserID, lot); err != nil {
			logger.Error("history lot not winner failed")
			return err
		}

		logger.Info("bet must be less than the price",
			zap.Uint("price", rule.price()),
			zap.Uint("value", curBet.Value),
		)
	}

	if err := prx.Next(); err != nil {
		logger.Error("next rule failed", zap.Error(err))
		return err
	}

	return nil
}
