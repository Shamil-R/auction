package rule

import (
	"gitlab/nefco/auction/process"

	"go.uber.org/zap"
)

type simple struct {
	*common
}

func newSimple(name string, interval process.Interval, config commonConfig) *simple {
	return &simple{newCommon(name, interval, config)}
}

func (rule *simple) Start(prx process.RuleProxy) error {
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

	if err := prx.ClearBetsBefore(lot.ID, rule.Interval().Start()); err != nil {
		logger.Error("clear bets failed", zap.Error(err))
		return err
	}

	lot.ClearBetsBefore(rule.Interval().Start())

	return nil
}
