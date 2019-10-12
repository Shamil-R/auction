package process

import (
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/core/object"
	"gitlab/nefco/auction/db"
	"gitlab/nefco/auction/errors"
	"gitlab/nefco/auction/service"
	"time"

	"go.uber.org/zap"
)

var (
	lotAccessDenied     = errors.Forbidden("Lot access denied")
	lotAlreadyBooked    = errors.BadRequest("Lot already booked")
	lotAlreadyConfirmed = errors.BadRequest("Lot already confirmed")
	lotAlreadyCompleted = errors.BadRequest("Lot already completed")
	confirmInfoInvalid  = errors.BadRequest("Confirm info invalid")
)

type Action interface {
	Executor() *core.User
	LotID() uint
	GetLot() *core.Lot
	SetLot(lot *core.Lot)
}

type PlaceBet interface {
	Action
	Value() uint
}

type ConfirmLot interface {
	Action
	Info() object.JSONData
}

type AcceptBet interface {
	Action
	BetID() uint64
}

type BackService interface {
	PostConfirmation(object.JSONData) error
	DeleteConfirmation() error
	SetDateBook(time.Time, *core.Bet) error
	ResetDateBook() error
}

type RuleProxy interface {
	BackService
	core.LotService
	core.BetService
	core.HistoryService
	End() time.Time
	ProcessLot() *core.Lot
	Run(rule Rule) error
	Next() error
	Complete() error
	Prolong(d time.Duration)
}

type ruleProxy struct {
	BackService
	core.LotService
	core.BetService
	core.HistoryService
	*process
	lot *core.Lot
}

func newRuleProxy(prc *process, tx *db.Tx,
	executor *core.User, lot *core.Lot) RuleProxy {
	return &ruleProxy{
		BackService:    service.NewBackService(prc.cfgBackService, executor, lot),
		LotService:     service.NewLotService(tx),
		BetService:     service.NewBetService(tx),
		HistoryService: service.NewHistoryService(tx),
		process:        prc,
		lot:            lot,
	}
}

func (prx *ruleProxy) ProcessLot() *core.Lot {
	return prx.lot
}

type Rule interface {
	Rule() string
	Config() interface{}
	Interval() Interval
	Start(prx RuleProxy) error
	Stop(prx RuleProxy) error
	PlaceBet(act PlaceBet, prx RuleProxy) error
	CancelBet(act Action, prx RuleProxy) error
	ConfirmLot(act ConfirmLot, prx RuleProxy) error
	AcceptBet(act AcceptBet, prx RuleProxy) error
	Sync(lot *core.Lot)
}

type Process interface {
	LotID() uint
	Stop() error
	PlaceBet(act PlaceBet) error
	CancelBet(act Action) error
	ConfirmLot(act ConfirmLot) error
	AcceptBet(act AcceptBet) error
	Sync(lot *core.Lot)
}

type ProcessService interface {
	Stop(prc Process) error
	LotChanged(lot *core.Lot)
	LotChangedWithReceiver(lot *core.Lot, receiver *core.User)
}

type process struct {
	lotID          uint
	rules          []Rule
	db             *db.DB
	prcSvc         ProcessService
	cfgBackService *service.ConfigBackService
	timeline       Timeline
	currentRule    Rule
	nextRule       Rule
	runner         *core.User
	logger         *zap.Logger
}

func New(
	executor *core.User, lot *core.Lot, rules []Rule,
	database *db.DB, prcSvc ProcessService,
	cfgBackService *service.ConfigBackService,
	tx *db.Tx, startRule Rule,
) (*process, error) {
	p := &process{
		lotID:          lot.ID,
		rules:          rules,
		db:             database,
		prcSvc:         prcSvc,
		cfgBackService: cfgBackService,
		timeline:       newTimeline(),
		runner:         executor,
		logger:         zap.L().Named("process").With(zap.Uint("lot_id", lot.ID)),
	}

	if err := p.create(executor, lot, tx, startRule); err != nil {
		return nil, err
	}

	return p, nil
}

func (prc *process) LotID() uint {
	return prc.lotID
}

func (prc *process) Stop() error {
	return prc.timeline.Stop()
}

func (prc *process) PlaceBet(act PlaceBet) error {
	logger := prc.logger.Named("place_bet").With(
		zap.Uint("executor_id", act.Executor().ID),
	)

	return prc.txRule(func(prx RuleProxy, tx *db.Tx) error {
		lot := prx.ProcessLot()

		if lot.BookedAt != nil {
			logger.Warn("lot already booked")
			return lotAlreadyBooked
		}

		if err := prc.currentRule.PlaceBet(act, prx); err != nil {
			return err
		}

		prc.Sync(lot)

		lot.UpdatePrice(act.Executor().ID)

		historySvc := service.NewHistoryService(tx)

		if err := historySvc.BetPlaced(act.Executor().ID, lot); err != nil {
			return err
		}

		return nil
	}, act)
}

func (prc *process) CancelBet(act Action) error {
	logger := prc.logger.Named("cancel_bet").With(
		zap.Uint("executor_id", act.Executor().ID),
	)

	return prc.txRule(func(prx RuleProxy, tx *db.Tx) error {
		lot := prx.ProcessLot()

		if lot.ConfirmedAt != nil {
			logger.Warn("lot already confirmed")
			return lotAlreadyConfirmed
		}

		if lot.CompletedAt != nil {
			logger.Warn("lot already completed")
			return lotAlreadyCompleted
		}

		if err := prc.currentRule.CancelBet(act, prx); err != nil {
			return err
		}

		prc.Sync(lot)

		lot.UpdatePrice(act.Executor().ID)

		historySvc := service.NewHistoryService(tx)

		if err := historySvc.BetCanceled(act.Executor().ID, lot); err != nil {
			return err
		}

		return nil
	}, act)
}

func (prc *process) ConfirmLot(act ConfirmLot) error {
	logger := prc.logger.Named("confirm_bet").With(
		zap.Uint("executor_id", act.Executor().ID),
	)

	return prc.txRule(func(prx RuleProxy, tx *db.Tx) error {
		lot := prx.ProcessLot()

		if lot.ConfirmedAt != nil {
			logger.Warn("lot already confirmed")
			return lotAlreadyConfirmed
		}

		ok, err := lot.Object.Type.CheckConfirm(act.Info())
		if err != nil {
			logger.Error("check confirm failed", zap.Error(err))
			return err
		}

		if !ok {
			logger.Warn("confirm info invalid")
			return confirmInfoInvalid
		}

		if err := prc.currentRule.ConfirmLot(act, prx); err != nil {
			return err
		}

		prc.Sync(lot)

		lot.UpdatePrice(act.Executor().ID)

		historySvc := service.NewHistoryService(tx)

		if err := historySvc.LotConfirmed(act.Executor().ID, lot); err != nil {
			return err
		}

		return nil
	}, act)
}

func (prc *process) AcceptBet(act AcceptBet) error {
	logger := prc.logger.Named("confirm_bet").With(
		zap.Uint("executor_id", act.Executor().ID),
	)

	return prc.txRule(func(prx RuleProxy, tx *db.Tx) error {
		lot := prx.ProcessLot()

		if lot.BookedAt != nil {
			logger.Warn("lot already booked")
			return lotAlreadyBooked
		}

		if err := prc.currentRule.AcceptBet(act, prx); err != nil {
			return err
		}

		prc.Sync(lot)

		lot.UpdatePrice(act.Executor().ID)

		historySvc := service.NewHistoryService(tx)

		if err := historySvc.LotBetAccept(act.Executor().ID, lot); err != nil {
			return err
		}

		return nil
	}, act)
}

func (prc *process) Sync(lot *core.Lot) {
	prc.currentRule.Sync(lot)
	lot.End = prc.End()
	lot.Rest = uint(lot.End.Sub(now()) / time.Second)
}

func (prc *process) End() time.Time {
	end := prc.timeline.End()
	if end.IsZero() {
		end = prc.currentRule.Interval().End()
	}
	if prc.nextRule != nil {
		end = prc.nextRule.Interval().End()
	}
	return end
}

func (prc *process) Run(rule Rule) error {
	prc.nextRule = rule
	prc.timeline.Stop()
	return nil
}

func (prc *process) Next() error {
	return prc.Run(prc.next())
}

func (prc *process) Complete() error {
	return prc.prcSvc.Stop(prc)
}

func (prc *process) Prolong(d time.Duration) {
	prc.timeline.Prolong(d)
}

func (prc *process) start(rule Rule, startHandler func(rule Rule) error) error {
	logger := prc.logger.Named("start")

	if rule == nil {
		return errors.New("rule can not be empty")
	}

	prc.currentRule = rule

	if err := startHandler(rule); err != nil {
		logger.Error("rule start failed", zap.Error(err))
		return err
	}

	go prc.run(rule)

	return nil
}

func (prc *process) create(executor *core.User, lot *core.Lot, tx *db.Tx,
	startRule Rule) error {
	logger := prc.logger.Named("restore")

	var rule Rule = startRule

	if rule == nil {
		rule = prc.next()
	}

	return prc.start(rule, func(rule Rule) error {
		if err := rule.Start(newRuleProxy(prc, tx, executor, lot)); err != nil {
			logger.Error("rule start failed", zap.Error(err))
			return err
		}
		return nil
	})
}

func (prc *process) next() Rule {
	now := now().Add(1 * time.Second)
	for _, rule := range prc.rules {
		if rule.Interval().Contains(now) {
			return rule
		}
	}
	return nil
}

func (prc *process) run(rule Rule) error {
	logger := prc.logger.Named("run")

	start := now()
	end := rule.Interval().End()

	logger.Debug("start",
		zap.String("rule", rule.Rule()),
		zap.Time("start", start),
		zap.Time("end", end),
	)

	complete := <-prc.timeline.Run(start, end)

	logger.Debug("complete",
		zap.String("rule", rule.Rule()),
		zap.Time("start", rule.Interval().Start()),
		zap.Time("end", prc.End()),
		zap.Bool("complete", complete),
	)

	if complete {
		if err := prc.ruleStop(rule); err != nil {
			logger.Error("rule stop failed", zap.Error(err))
		}
	}

	if prc.nextRule != nil {
		if err := prc.start(prc.nextRule, prc.ruleStart); err != nil {
			logger.Error("start next rule failed", zap.Error(err))
		}
		prc.nextRule = nil
	}

	return nil
}

func (prc *process) ruleStart(rule Rule) error {
	return prc.txRule(func(prx RuleProxy, tx *db.Tx) error {
		if err := rule.Start(prx); err != nil {
			return err
		}

		lot := prx.ProcessLot()

		prc.Sync(lot)

		prc.prcSvc.LotChanged(lot)

		return nil
	}, nil)
}

func (prc *process) ruleStop(rule Rule) error {
	return prc.txRule(func(prx RuleProxy, tx *db.Tx) error {
		lot := prx.ProcessLot()

		prc.Sync(lot)

		lot.UpdatePrice(lot.UserID)

		if err := rule.Stop(prx); err != nil {
			return err
		}

		// if lot == nil {
		// 	return nil
		// }

		historySvc := service.NewHistoryService(tx)

		if lot.CurrentBet() != nil && lot.CompletedAt != nil {
			if err := historySvc.LotCompleted(lot.CurrentBet().UserID, lot); err != nil {
				return err
			}
		} else if lot.CurrentBet() != nil && lot.BookedAt != nil {
			if err := historySvc.LotBooked(lot.CurrentBet().UserID, lot); err != nil {
				return err
			}
		}

		prc.prcSvc.LotChanged(lot)

		return nil
	}, nil)
}

func (prc *process) tx(handler func(tx *db.Tx) error) error {
	logger := prc.logger.Named("tx")

	tx, err := prc.db.Begin()
	if err != nil {
		logger.Error("transaction begin failed", zap.Error(err))
		return err
	}

	if err := handler(tx); err != nil {
		if err := tx.Rollback(); err != nil {
			logger.Error("transaction rollback failed", zap.Error(err))
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		logger.Error("transaction commit failed", zap.Error(err))
		return err
	}

	return nil
}

func (prc *process) txRule(
	handler func(RuleProxy, *db.Tx) error, act Action,
) error {
	logger := prc.logger.Named("txRule")

	return prc.tx(func(tx *db.Tx) error {
		svc := service.NewLotService(tx)

		lot, err := svc.Lot(prc.lotID, false)
		if err != nil {
			logger.Error("get lot failed", zap.Error(err))
			return err
		}

		executor := prc.runner

		if act != nil {
			if !act.Executor().Check(lot) {
				return lotAccessDenied
			}
			act.SetLot(lot)

			executor = act.Executor()
		}

		if err := handler(newRuleProxy(prc, tx, executor, lot), tx); err != nil {
			logger.Error("handler failed", zap.Error(err))
			return err
		}

		return nil
	})
}
