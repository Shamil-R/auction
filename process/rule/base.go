package rule

// Rule structure:
// - base
//   - wait
//   - pick
//   - confirm
//   - common
//     - hot
//     - simple
//     	 - normal
//     	 - extra

import (
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/errors"
	"gitlab/nefco/auction/process"

	"go.uber.org/zap"
)

var (
	placeBetDisabled      = errors.BadRequest("Place bet disabled")
	cancelBetDisabled     = errors.BadRequest("Cancel bet disabled")
	confirmDisabled       = errors.BadRequest("Confirm disabled")
	betMustLessCurrentBet = errors.BadRequest("Bet must be less than the current bet")
	betInvalid            = errors.BadRequest("Bet invalid")
	betStepInvalid        = errors.BadRequest("Bet step invalid")
	betAlreadyExist       = errors.BadRequest("Bet is already made by the current user")
	betNotFound           = errors.BadRequest("Bet not found")
	betOtherUser          = errors.BadRequest("Bid made by another user")
)

type base struct {
	name     string
	interval process.Interval
	logger   *zap.Logger
}

func newBase(name string, interval process.Interval) *base {
	return &base{
		name:     name,
		interval: interval,
		logger:   zap.L().Named("rules").Named(name),
	}
}

func (rule *base) Rule() string {
	return rule.name
}

func (rule *base) Config() interface{} {
	return nil
}

func (rule *base) Sync(lot *core.Lot) {
	lot.Rule = rule.name
}

func (rule *base) Interval() process.Interval {
	return rule.interval
}

func (rule *base) Start(prx process.RuleProxy) error {
	return nil
}

func (rule *base) Stop(prx process.RuleProxy) error {
	logger := rule.logger.Named("stop")

	if err := prx.Next(); err != nil {
		logger.Error("next rule failed", zap.Error(err))
		return err
	}

	return nil
}

func (rule *base) PlaceBet(act process.PlaceBet, prx process.RuleProxy) error {
	return placeBetDisabled
}

func (rule *base) CancelBet(act process.Action, prx process.RuleProxy) error {
	return cancelBetDisabled
}

func (rule *base) ConfirmLot(act process.ConfirmLot, prx process.RuleProxy) error {
	return confirmDisabled
}

func (rule *base) AcceptBet(act process.AcceptBet, prx process.RuleProxy) error {
	return confirmDisabled
}
