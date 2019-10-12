package rule

import (
	"gitlab/nefco/auction/process"

	"go.uber.org/zap"
)

const Wait = "wait"

type wait struct {
	*base
}

func newWait(interval process.Interval) *wait {
	return &wait{newBase(Wait, interval)}
}

func (rule *wait) Start(prx process.RuleProxy) error {
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
