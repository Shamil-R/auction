package rule

import (
	"gitlab/nefco/auction/process"
	"time"

	"go.uber.org/zap"
)

const Confirm = "confirm"

type confirm struct {
	*base
}

func NewConfirm(duration time.Duration) (*confirm, error) {
	n := now()

	interval, err := process.NewInterval(n, duration)
	if err != nil {
		return nil, err
	}

	return &confirm{newBase(Confirm, interval)}, nil
}

func (rule *confirm) Stop(prx process.RuleProxy) error {
	logger := rule.logger.Named("stop")

	lot := prx.ProcessLot()

	lot.BookedAt = nil

	if err := prx.SaveLot(lot); err != nil {
		logger.Error("save lot failed", zap.Error(err))
		return err
	}

	if err := prx.ClearBets(lot.ID); err != nil {
		logger.Error("clear bets failed", zap.Error(err))
		return err
	}

	lot.ClearBets()

	if err := prx.ResetDateBook(); err != nil {
		logger.Error("reset date book failed", zap.Error(err))
		return err
	}

	if err := prx.LotNoConfirm(lot.UserID, lot); err != nil {
		logger.Error("history lot not confirm failed")
		return err
	}

	if err := prx.Next(); err != nil {
		logger.Error("next rule failed", zap.Error(err))
		return err
	}

	return nil
}

func (rule *confirm) CancelBet(act process.Action, prx process.RuleProxy) error {
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

	lot.BookedAt = nil

	if err := prx.SaveLot(lot); err != nil {
		logger.Error("save lot failed", zap.Error(err))
		return err
	}

	if err := prx.ClearBets(lot.ID); err != nil {
		logger.Error("clear bets failed", zap.Error(err))
		return err
	}

	lot.ClearBets()

	if err := prx.ResetDateBook(); err != nil {
		logger.Error("reset date book failed", zap.Error(err))
		return err
	}

	if err := prx.Next(); err != nil {
		logger.Error("next rule failed", zap.Error(err))
		return err
	}

	return nil
}

func (rule *confirm) ConfirmLot(act process.ConfirmLot, prx process.RuleProxy) error {
	logger := rule.logger.Named("confirm")

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

	if err := prx.PostConfirmation(act.Info()); err != nil {
		logger.Error("back return error", zap.Error(err))
		return err
	}

	n := now()

	lot.ConfirmedAt = &n
	lot.Confirm = act.Info()

	if err := prx.SaveLot(lot); err != nil {
		logger.Error("save lot failed", zap.Error(err))
		return err
	}

	if err := prx.Complete(); err != nil {
		logger.Error("complete failed", zap.Error(err))
		return err
	}

	return nil
}
