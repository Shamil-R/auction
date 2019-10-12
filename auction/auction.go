package auction

import (
	"encoding/json"
	"fmt"
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/core/object"
	"gitlab/nefco/auction/core/object/json_fileds"
	"gitlab/nefco/auction/db"
	"gitlab/nefco/auction/errors"
	"gitlab/nefco/auction/interfaces"
	"gitlab/nefco/auction/process"
	"gitlab/nefco/auction/process/rule"
	"gitlab/nefco/auction/service"
	"time"

	"go.uber.org/zap"
)

var (
	groupNotFound           = errors.NotFound("Group not found")
	groupAlreadyExist       = errors.BadRequest("Group already exist")
	userAlreadyExist        = errors.BadRequest("User already exist")
	lotAlreadyExist         = errors.BadRequest("Lot already exist")
	lotTypeAccessDenied     = errors.Forbidden("Lot type access denied")
	processNotFound         = errors.BadRequest("Process not found")
	lotNotBooked            = errors.BadRequest("Lot not booked")
	lotNotConfirmed         = errors.BadRequest("Lot not confirmed")
	lotAlreadyConfirmed     = errors.BadRequest("Lot already confirmed")
	confirmInfoInvalid      = errors.BadRequest("Confirm info invalid")
	lotNotCompleted         = errors.BadRequest("Lot not completed")
	completeInfoInvalid     = errors.BadRequest("Complete info invalid")
	lotCompleteAccessDenied = errors.BadRequest("Lot complete access denied")
	lotAlreadyCompleted     = errors.BadRequest("Lot already completed")
	betNotFound             = errors.BadRequest("Bet not found")
	betOtherUser            = errors.BadRequest("Bid made by another user")
	userBlocked             = errors.BadRequest("User bloked")
	actNotFound 			= errors.NotFound("act not found")
)

type Action interface {
	Executor() *core.User
}

type ActionUser interface {
	Action
	UserID() uint
}

type ActionLot interface {
	Action
	LotID() uint
}

type Groups interface {
	Action
	core.ObjectTypeFilter
	SetGroups(groups []*core.Group)
}

type CreateGroup interface {
	Action
	GetGroup() *core.Group
}

type Users interface {
	Action
	SetUsers(users []*core.User)
}

type GetUser interface {
	ActionUser
	SetUser(user *core.User)
}

type CreateUser interface {
	Action
	GetUser() *core.User
}

type EditUser interface {
	Action
	GetUser() *core.UserInfo
}

type AddUserGroup interface {
	Action
	core.UserGroup
}

type DeleteUserGroup interface {
	Action
	core.UserGroup
}

type Lots interface {
	core.LotsFilter
	SetLots(lots []*core.Lot)
}

type CreateLot interface {
	Action
	GetLot() *core.Lot
}

type GetLot interface {
	ActionLot
	SetLot(lot *core.Lot)
}

type UpdateLot interface {
	ActionLot
	GetLot() *core.Lot
	SetLot(lot *core.Lot)
}

type CompleteLot interface {
	ActionLot
	Info() object.JSONData
}

type History interface {
	ActionLot
	SetHistory(histories []*core.History)
}

type SendFeedback interface {
	Message() string
}

type GetBookingWindow interface {
	ActionLot
	SetLoad(ld []core.LoadDate)
}

type NotifyWriter interface {
	LotAdded(lot *core.Lot)
	LotChanged(lot core.Lot)
	LotChangedWithReceiver(lot *core.Lot, receiver *core.User)
	LotDeleted(lot *core.Lot)
}

type FeedbackService interface {
	Send(message string) error
}

type Auction interface {
	Restore(executor *core.User) error
	Groups(act Groups) error
	CreateGroup(act CreateGroup) error
	Users(act Users) error
	User(username string) (*core.User, error)
	CreateUser(act CreateUser) error
	GetUser(act GetUser) error
	EditUser(act EditUser) error
	AddUserGroup(act AddUserGroup) error
	DeleteUserGroup(act DeleteUserGroup) error
	BlockUser(userID uint) error
	UnblockUser(userID uint) error
	Lots(act Lots) error
	CreateLot(act CreateLot) error
	GetLot(act GetLot) error
	UpdateLot(act UpdateLot) error
	DeleteLot(act ActionLot) error
	History(act History) error
	PlaceBet(act process.PlaceBet) error
	CancelBet(act process.Action) error
	ConfirmLot(act process.ConfirmLot) error
	DeleteConfirmation(act ActionLot) error
	CompleteLot(act CompleteLot) error
	AcceptBet(act process.AcceptBet) error
	SendFeedback(act SendFeedback) error
	GetBookingWindow(act GetBookingWindow) error
	UpdateLotAct(act interfaces.EditActCommander) error
	AutoBookingLot(act interfaces.AutoBookingCommander) error
	AllowChangeAct(act interfaces.AllowChangeActCommander) error
}

type auction struct {
	config         *Config
	db             *db.DB
	notify         NotifyWriter
	feedbackSvc    FeedbackService
	cfgBackService *service.ConfigBackService
	pp             map[uint]process.Process
	logger         *zap.Logger
}

func New(config *Config, db *db.DB, notify NotifyWriter,
	feedbackSvc FeedbackService,
	cfgBackService *service.ConfigBackService) *auction {
	return &auction{
		config:         config,
		db:             db,
		notify:         notify,
		feedbackSvc:    feedbackSvc,
		cfgBackService: cfgBackService,
		pp:             make(map[uint]process.Process),
		logger:         zap.L().Named("auction"),
	}
}

func (auc *auction) Restore(executor *core.User) error {
	logger := auc.logger.Named("restore")

	return auc.tx(func(tx *db.Tx) error {
		svc := service.NewLotService(tx)

		lots, err := svc.Lots(nil)
		if err != nil {
			logger.Error("get lots failed", zap.Error(err))
			return err
		}

		for _, lot := range lots {
			rules, err := rule.Rules(lot.Rules, auc.config.RuleConfig)
			if err != nil {
				logger.Error("default rules failed", zap.Error(err))
				return err
			}

			var startRule process.Rule

			n := process.Now()

			if lot.BookedAt != nil {
				d := n.Sub(*lot.BookedAt)
				if d < rule.DefaultConfirmDuration {
					startRule, err = rule.NewConfirm(rule.DefaultConfirmDuration - d)
					if err != nil {
						logger.Error("confirm rule failed", zap.Error(err))
						return err
					}
				}
			}

			process, err := process.New(executor, lot, rules, auc.db,
				auc, auc.cfgBackService, tx, startRule)
			if err != nil {
				logger.Error("process failed", zap.Error(err))
				return err
			}

			auc.pp[lot.ID] = process
		}

		logger.Info("restore lots", zap.Int("count", len(lots)))

		return nil
	})
}

func (auc *auction) Stop(process process.Process) error {
	if _, ok := auc.pp[process.LotID()]; !ok {
		return processNotFound
	}
	if err := process.Stop(); err != nil {
		return err
	}
	delete(auc.pp, process.LotID())
	return nil
}

func (auc *auction) Users(act Users) error {
	logger := auc.logger.Named("users")

	return auc.tx(func(tx *db.Tx) error {
		svc := service.NewUserService(tx)

		users, err := svc.Users()
		if err != nil {
			logger.Error("get users failed", zap.Error(err))
			return err
		}

		act.SetUsers(users)

		return nil
	})
}

func (auc *auction) LotChanged(lot *core.Lot) {
	syncLots(auc.pp, lot)
	auc.notify.LotChanged(*lot)
}

func (auc *auction) LotChangedWithReceiver(lot *core.Lot, receiver *core.User) {
	syncLots(auc.pp, lot)
	auc.notify.LotChangedWithReceiver(lot, receiver)
}

func (auc *auction) LotDeleted(lot *core.Lot) {
	syncLots(auc.pp, lot)
	auc.notify.LotDeleted(lot)
}

func (auc *auction) Groups(act Groups) error {
	logger := auc.logger.Named("groups")

	return auc.tx(func(tx *db.Tx) error {
		svc := service.NewGroupService(tx)

		groups, err := svc.Groups(act)
		if err != nil {
			logger.Error("get groups failed", zap.Error(err))
			return err
		}

		act.SetGroups(groups)

		return nil
	})
}

func (auc *auction) CreateGroup(act CreateGroup) error {
	logger := auc.logger.Named("create_group")

	return auc.tx(func(tx *db.Tx) error {
		svc := service.NewGroupService(tx)

		newGroup := act.GetGroup()

		group, err := svc.GroupByKey(newGroup.Key)
		if err != nil {
			logger.Error("get group failed", zap.Error(err))
			return err
		}

		if group != nil {
			return groupAlreadyExist
		}

		if err := svc.CreateGroup(act.GetGroup()); err != nil {
			logger.Error("create group failed", zap.Error(err))
			return err
		}

		return nil
	})
}

func (auc *auction) User(username string) (*core.User, error) {
	logger := auc.logger.Named("user")

	tx, err := auc.db.Begin()
	if err != nil {
		logger.Error("transaction begin failed", zap.Error(err))
		return nil, err
	}

	svc := service.NewUserService(tx)

	user, err := svc.UserByUsername(username)
	if err != nil {
		logger.Warn("get user failed", zap.Error(err))
		if err := tx.Rollback(); err != nil {
			logger.Error("transaction rollback failed", zap.Error(err))
		}
		return nil, err
	}

	groups, err := svc.GroupsByUser(user.ID)
	if err != nil {
		logger.Error("get groups by user failed", zap.Error(err))
		if err := tx.Rollback(); err != nil {
			logger.Error("transaction rollback failed", zap.Error(err))
		}
		return nil, err
	}

	user.Groups = groups

	if err := tx.Commit(); err != nil {
		logger.Error("transaction commit failed", zap.Error(err))
		return nil, err
	}

	return user, nil
}

func (auc *auction) CreateUser(act CreateUser) error {
	logger := auc.logger.Named("create_user")

	user := act.GetUser()

	password, err := core.HashPassword(user.Password)
	if err != nil {
		logger.Error("invalid password", zap.Error(err))
		return err
	}

	user.Password = password

	return auc.tx(func(tx *db.Tx) error {
		svc := service.NewUserService(tx)

		_, err := svc.UserByUsername(user.Username)
		if err != nil && err != core.UserNotFound {
			logger.Error("get user failed", zap.Error(err))
			return err
		}

		if err == nil {
			return userAlreadyExist
		}

		if err := svc.CreateUser(user); err != nil {
			logger.Error("create user failed", zap.Error(err))
			return err
		}

		user.Password = ""

		return nil
	})
}

func (auc *auction) EditUser(act EditUser) error {
	logger := auc.logger.Named("edit_user")

	userCredentials := act.GetUser()

	user := core.User{
		ID:       act.Executor().ID,
		UserInfo: *userCredentials,
	}

	return auc.tx(func(tx *db.Tx) error {
		svc := service.NewUserService(tx)

		err := svc.SaveUser(&user)
		if err != nil {
			logger.Error("save user failed", zap.Error(err))
			return err
		}

		return nil
	})
}

func (auc *auction) GetUser(act GetUser) error {
	logger := auc.logger.Named("get_user")

	return auc.tx(func(tx *db.Tx) error {
		svc := service.NewUserService(tx)

		user, err := svc.User(act.UserID())
		if err != nil {
			logger.Error("get user failed", zap.Error(err))
			return err
		}

		groups, err := svc.GroupsByUser(user.ID)
		if err != nil {
			logger.Error("get groups failed", zap.Error(err))
			return err
		}

		user.Groups = groups

		act.SetUser(user)

		return nil
	})
}

func (auc *auction) AddUserGroup(act AddUserGroup) error {
	logger := auc.logger.Named("add_user_group")

	return auc.tx(func(tx *db.Tx) error {
		svc := service.NewUserService(tx)

		if err := svc.AddGroup(act); err != nil {
			logger.Error("add group failed", zap.Error(err))
			return err
		}

		return nil
	})
}

func (auc *auction) DeleteUserGroup(act DeleteUserGroup) error {
	logger := auc.logger.Named("delete_user_group")

	return auc.tx(func(tx *db.Tx) error {
		svc := service.NewUserService(tx)

		if err := svc.DeleteGroup(act); err != nil {
			logger.Error("delete group failed", zap.Error(err))
			return err
		}

		return nil
	})
}

func (auc *auction) BlockUser(userID uint) error {
	logger := auc.logger.Named("block_user")
	return auc.tx(func(tx *db.Tx) error {
		srv := service.NewUserService(tx)

		err := srv.BlockUser(userID)
		if err != nil {
			logger.Error("block user failed", zap.Error(err))
			return err
		}
		return nil
	})
}

func (auc *auction) UnblockUser(userID uint) error {
	logger := auc.logger.Named("unblock_user")
	return auc.tx(func(tx *db.Tx) error {
		srv := service.NewUserService(tx)

		err := srv.UnblockUser(userID)
		if err != nil {
			logger.Error("unblock user failed", zap.Error(err))
			return err
		}
		return nil
	})
}

func (auc *auction) Lots(act Lots) error {
	logger := auc.logger.Named("lots")

	return auc.tx(func(tx *db.Tx) error {
		svc := service.NewLotService(tx)

		lots, err := svc.Lots(act)
		if err != nil {
			logger.Error("get lots failed", zap.Error(err))
			return err
		}

		syncLots(auc.pp, lots...)

		restLots := make([]*core.Lot, 0, len(lots))
		for _, lot := range lots {
			if !act.Executor().CheckFilter(lot) {
				continue
			}
			lot.UpdatePrice(act.Executor().ID)
			restLots = append(restLots, lot)
		}

		act.SetLots(restLots)

		return nil
	})
}

func (auc *auction) CreateLot(act CreateLot) error {
	logger := auc.logger.Named("create_lot")

	return auc.tx(func(tx *db.Tx) error {
		lot := act.GetLot()

		if !act.Executor().Check(lot) {
			logger.Warn("lot type access denied")
			return lotTypeAccessDenied
		}

		rules, err := rule.Rules(lot.Rules, auc.config.RuleConfig)
		if err != nil {
			logger.Error("default rules failed", zap.Error(err))
			return err
		}

		lot.User = act.Executor()
		lot.UserID = lot.User.ID

		svcLot := service.NewLotService(tx)

		exLot, err := svcLot.LotByObject(lot.GroupKey, lot.ObjectID)
		if err != nil && err != core.LotNotFound {
			logger.Error("get lot by object failed", zap.Error(err))
			return err
		}

		isManualBooked := lot.BookedAt != nil && lot.Bet != nil

		lot.ManualBooked = isManualBooked

		if err == nil {
			lot.ID = exLot.ID
			lot.DeletedAt = exLot.DeletedAt
			if _, ok := auc.pp[lot.ID]; ok || (isManualBooked && lot.DeletedAt == nil) {
				logger.Warn("lot already exist")
				return lotAlreadyExist
			}

			lot.DeletedAt = nil

			if err := svcLot.SaveLot(lot); err != nil {
				logger.Error("save lot failed", zap.Error(err))
				return err
			}
		} else {
			if err := svcLot.CreateLot(lot); err != nil {
				logger.Error("create lot failed", zap.Error(err))
				return err
			}
		}

		if isManualBooked {
			bet := lot.Bet
			bet.Winner = true
			bet.LotID = lot.ID

			svcBet := service.NewBetService(tx)

			if err := svcBet.CreateBet(bet); err != nil {
				logger.Error("create bet failed", zap.Error(err))
				return err
			}
		} else {
			process, err := process.New(act.Executor(), lot, rules, auc.db, auc,
				auc.cfgBackService, tx, nil)
			if err != nil {
				logger.Error("process failed", zap.Error(err))
				return err
			}

			auc.pp[lot.ID] = process

			syncLots(auc.pp, lot)
		}

		svcHistory := service.NewHistoryService(tx)

		if err := svcHistory.LotAdded(lot.User.ID, lot); err != nil {
			logger.Error("history lot added failed", zap.Error(err))
			return err
		}

		if !isManualBooked {
			auc.notify.LotAdded(lot)
		}

		return nil
	})
}

func (auc *auction) GetLot(act GetLot) error {
	logger := auc.logger.Named("get_lot")

	return auc.tx(func(tx *db.Tx) error {
		svc := service.NewLotService(tx)

		lot, err := svc.Lot(act.LotID(), false)
		if err != nil {
			logger.Error("get lot failed", zap.Error(err))
			return err
		}

		if !act.Executor().Check(lot) {
			logger.Warn("lot type access denied")
			return lotTypeAccessDenied
		}

		syncLots(auc.pp, lot)

		lot.UpdatePrice(act.Executor().ID)

		if lot.Complete != nil {
			complete := &json_fileds.Complete{}
			if err := json.Unmarshal(lot.Complete, complete); err != nil {
				logger.Error(fmt.Sprintf("unmarshal lot.complete data failed, lot_id = %d", lot.ID), zap.Error(err))
				return errors.NewError("complete value incorrect!", 500)
			}

			// если нету поля, значит запрещено редактировать
			var allowChange int
			if complete.AllowChange == nil {
				allowChange = 0
				complete.AllowChange = &allowChange
			}

			b, err := json.Marshal(complete)
			if err != nil {
				logger.Error("marhsal allow_change failed", zap.Error(err))
				return err
			}

			lot.Complete = b
		}

		act.SetLot(lot)
		return nil
	})
}

func (auc *auction) UpdateLot(act UpdateLot) error {
	logger := auc.logger.Named("update_lot")

	return auc.tx(func(tx *db.Tx) error {
		svc := service.NewLotService(tx)

		lot, err := svc.Lot(act.LotID(), false)
		if err != nil {
			logger.Error("get lot failed", zap.Error(err))
			return err
		}

		if !act.Executor().Check(lot) {
			logger.Warn("lot type access denied")
			return lotTypeAccessDenied
		}

		lot.Rules = act.GetLot().Rules
		lot.Object = act.GetLot().Object
		lot.CompletedAt = act.GetLot().CompletedAt
		lot.Urgent = act.GetLot().Urgent

		if err := svc.SaveLot(lot); err != nil {
			logger.Error("save lot failed", zap.Error(err))
			return err
		}

		svcHistory := service.NewHistoryService(tx)

		if err := svcHistory.LotUpdated(lot.User.ID, lot); err != nil {
			logger.Error("history lot updated failed", zap.Error(err))
			return err
		}

		act.SetLot(lot)

		auc.LotChanged(lot)

		return nil
	})
}

func (auc *auction) DeleteLot(act ActionLot) error {
	logger := auc.logger.Named("delete_lot")

	return auc.tx(func(tx *db.Tx) error {
		svc := service.NewLotService(tx)

		lot, err := svc.Lot(act.LotID(), false)
		if err != nil {
			logger.Error("get lot failed", zap.Error(err))
			return err
		}

		if !act.Executor().Check(lot) {
			logger.Warn("lot type access denied")
			return lotTypeAccessDenied
		}

		if err := svc.DeleteLot(lot); err != nil {
			logger.Error("delete lot failed", zap.Error(err))
			return err
		}

		svcBet := service.NewBetService(tx)

		if err := svcBet.ClearBets(lot.ID); err != nil {
			logger.Error("clear bets failed", zap.Error(err))
			return err
		}

		if process, ok := auc.pp[lot.ID]; ok {
			if err := process.Stop(); err != nil {
				logger.Error("process stop failed", zap.Error(err))
				return err
			}
			delete(auc.pp, lot.ID)
		}

		svcHistory := service.NewHistoryService(tx)

		if err := svcHistory.LotDeleted(lot.User.ID, lot); err != nil {
			logger.Error("history lot deleted failed", zap.Error(err))
			return err
		}

		auc.LotDeleted(lot)

		return nil
	})
}

func (auc *auction) History(act History) error {
	logger := auc.logger.Named("history")

	return auc.tx(func(tx *db.Tx) error {
		svc := service.NewHistoryService(tx)

		histories, err := svc.History(act.LotID())
		if err != nil {
			logger.Error("get history failed", zap.Error(err))
			return err
		}

		act.SetHistory(histories)

		return nil
	})
}

func (auc *auction) PlaceBet(act process.PlaceBet) error {
	logger := auc.logger.Named("place_bet")
	user, err := auc.User(act.Executor().Username)
	if err != nil {
		logger.Error("get user failed")
		return err
	}

	if user.Blocked {
		logger.Warn("user blocked")
		return userBlocked
	}
	if process, ok := auc.pp[act.LotID()]; ok {
		if err := process.PlaceBet(act); err != nil {
			return err
		}
		auc.LotChanged(act.GetLot())
		return nil
	}
	return processNotFound
}

func (auc *auction) CancelBet(act process.Action) error {
	logger := auc.logger.Named("cancel_lot")

	if process, ok := auc.pp[act.LotID()]; ok {
		if err := process.CancelBet(act); err != nil {
			return err
		}
		auc.LotChanged(act.GetLot())
		return nil
	}

	return auc.tx(func(tx *db.Tx) error {
		svcLot := service.NewLotService(tx)

		lot, err := svcLot.Lot(act.LotID(), false)
		if err != nil {
			logger.Error("get lot failed", zap.Error(err))
			return err
		}

		if lot.ConfirmedAt != nil {
			logger.Warn("lot already confirmed")
			return lotAlreadyConfirmed
		}

		if lot.CompletedAt != nil {
			logger.Warn("lot already completed")
			return lotAlreadyCompleted
		}

		curBet := lot.CurrentBet()
		if curBet == nil {
			logger.Warn("bet not found")
			return betNotFound
		}

		if curBet.UserID != act.Executor().ID {
			logger.Warn("bet other user")
			return betOtherUser
		}

		svcBet := service.NewBetService(tx)

		if err := svcBet.DeleteBet(curBet); err != nil {
			logger.Error("delete bet failed", zap.Error(err))
			return err
		}

		if lot.BookedAt != nil {
			lot.BookedAt = nil
			lot.ManualBooked = false

			if err := svcLot.SaveLot(lot); err != nil {
				logger.Error("save lot failed", zap.Error(err))
				return err
			}

			svcBack := service.NewBackService(auc.cfgBackService, act.Executor(), lot)

			if err := svcBack.ResetDateBook(); err != nil {
				logger.Error("reset date book failed", zap.Error(err))
				return err
			}
		}

		rules, err := rule.Rules(lot.Rules, auc.config.RuleConfig)
		if err != nil {
			logger.Error("rules failed", zap.Error(err))
			return err
		}

		process, err := process.New(act.Executor(), lot, rules, auc.db, auc,
			auc.cfgBackService, tx, nil)
		if err != nil {
			logger.Error("process failed", zap.Error(err))
			return err
		}

		auc.pp[lot.ID] = process

		syncLots(auc.pp, lot)

		historySvc := service.NewHistoryService(tx)

		if err := historySvc.BetCanceled(act.Executor().ID, lot); err != nil {
			return err
		}

		auc.LotChanged(lot)

		return nil
	})
}

func (auc *auction) ConfirmLot(act process.ConfirmLot) error {
	logger := auc.logger.Named("confirm_lot")

	if process, ok := auc.pp[act.LotID()]; ok {
		if err := process.ConfirmLot(act); err != nil {
			return err
		}
		auc.LotChangedWithReceiver(act.GetLot(), act.Executor())
		return nil
	}

	return auc.tx(func(tx *db.Tx) error {
		svcLot := service.NewLotService(tx)

		lot, err := svcLot.Lot(act.LotID(), false)
		if err != nil {
			logger.Error("get lot failed", zap.Error(err))
			return err
		}

		if lot.BookedAt == nil {
			logger.Warn("lot not booked")
			return lotNotBooked
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

		curBet := lot.CurrentBet()
		if curBet == nil {
			logger.Warn("bet not found")
			return betNotFound
		}

		if curBet.UserID != act.Executor().ID {
			logger.Warn("bet other user")
			return betOtherUser
		}

		backSvc := service.NewBackService(auc.cfgBackService, act.Executor(), lot)
		if err := backSvc.PostConfirmation(act.Info()); err != nil {
			logger.Error("post confirmation failed", zap.Error(err))
			return err
		}

		isConfirmEdit := lot.ConfirmedAt != nil

		n := process.Now()

		if !isConfirmEdit {
			lot.ConfirmedAt = &n
		}

		lot.Confirm = act.Info()

		if err := svcLot.SaveLot(lot); err != nil {
			logger.Error("save lot failed", zap.Error(err))
			return err
		}

		lot.UpdatePrice(act.Executor().ID)

		historySvc := service.NewHistoryService(tx)

		if isConfirmEdit {
			err := historySvc.LotConfirmEdited(act.Executor().ID, lot)
			if err != nil {
				logger.Error("history lot confirm edited failed",
					zap.Error(err),
				)
				return err
			}
		} else {
			err := historySvc.LotConfirmed(act.Executor().ID, lot)
			if err != nil {
				logger.Error("history lot confirmed failed", zap.Error(err))
				return err
			}
		}

		auc.LotChangedWithReceiver(lot, act.Executor())

		return nil
	})
}

func (auc *auction) DeleteConfirmation(act ActionLot) error {
	logger := auc.logger.Named("delete_confirmation")

	return auc.tx(func(tx *db.Tx) error {
		svc := service.NewLotService(tx)

		lot, err := svc.Lot(act.LotID(), false)
		if err != nil {
			logger.Error("get lot failed", zap.Error(err))
			return err
		}

		if !act.Executor().Check(lot) {
			logger.Warn("lot type access denied")
			return lotTypeAccessDenied
		}

		if lot.ConfirmedAt == nil {
			logger.Warn("lot not confirmed")
			return lotNotConfirmed
		}

		if lot.CompletedAt != nil {
			logger.Warn("lot already completed")
			return lotAlreadyCompleted
		}

		backSvc := service.NewBackService(auc.cfgBackService, act.Executor(), lot)
		if err := backSvc.DeleteConfirmation(); err != nil {
			logger.Error("back service return error", zap.Error(err))
			return err
		}

		svcBet := service.NewBetService(tx)

		if err := svcBet.ClearBets(lot.ID); err != nil {
			logger.Error("clear bets failed", zap.Error(err))
			return err
		}

		lot.BookedAt = nil
		lot.ManualBooked = false
		lot.ConfirmedAt = nil
		lot.Confirm = nil

		if err := svc.SaveLot(lot); err != nil {
			logger.Error("save lot failed", zap.Error(err))
			return err
		}

		rules, err := rule.Rules(lot.Rules, auc.config.RuleConfig)
		if err != nil {
			logger.Error("rules failed", zap.Error(err))
			return err
		}

		process, err := process.New(act.Executor(), lot, rules, auc.db, auc,
			auc.cfgBackService, tx, nil)
		if err != nil {
			logger.Error("process failed", zap.Error(err))
			return err
		}

		auc.pp[lot.ID] = process

		syncLots(auc.pp, lot)

		lot.UpdatePrice(act.Executor().ID)

		historySvc := service.NewHistoryService(tx)

		if err := historySvc.LotDeleteConfirmation(act.Executor().ID, lot); err != nil {
			logger.Error("history lot delete confirmation failed", zap.Error(err))
			return err
		}

		auc.LotChanged(lot)

		return err
	})
}

func (auc *auction) CompleteLot(act CompleteLot) error {
	logger := auc.logger.Named("complete_lot")

	return auc.tx(func(tx *db.Tx) error {
		svc := service.NewLotService(tx)

		lot, err := svc.Lot(act.LotID(), false)
		if err != nil {
			logger.Error("get lot failed", zap.Error(err))
			return err
		}

		if !act.Executor().Check(lot) {
			logger.Warn("lot type access denied")
			return lotTypeAccessDenied
		}

		curBet := lot.CurrentBet()
		if curBet == nil {
			logger.Warn("bet not found")
			return betNotFound
		}

		if curBet.UserID != act.Executor().ID && lot.UserID != act.Executor().ID {
			logger.Warn("lot complete access denied")
			return lotCompleteAccessDenied
		}

		if lot.CompletedAt == nil {
			return lotNotCompleted
		}

		ok, err := lot.Object.Type.CheckComplete(act.Info())
		if err != nil {
			logger.Error("check complete failed", zap.Error(err))
			return err
		}

		if !ok {
			logger.Warn("complete info invalid")
			return completeInfoInvalid
		}

		lot.Complete = act.Info()

		if err := svc.SaveLot(lot); err != nil {
			logger.Error("save lot failed", zap.Error(err))
			return nil
		}

		if curBet.UserID == act.Executor().ID {
			bcksvc := service.NewBackService(auc.cfgBackService, act.Executor(), lot)
			if err := bcksvc.ActDataSync(act.Info()); err != nil {
				logger.Error("send act data failed", zap.Error(err))
				return err
			}
		}

		return nil
	})
}

func (auc *auction) AcceptBet(act process.AcceptBet) error {
	if process, ok := auc.pp[act.LotID()]; ok {
		if err := process.AcceptBet(act); err != nil {
			return err
		}
		auc.LotChanged(act.GetLot())
		return nil
	}
	return processNotFound
}

func (auc *auction) SendFeedback(act SendFeedback) error {
	return auc.feedbackSvc.Send(act.Message())
}

func (auc *auction) GetBookingWindow(act GetBookingWindow) error {
	logger := auc.logger.Named("get_booking_window")

	return auc.tx(func(tx *db.Tx) error {
		svcLot := service.NewLotService(tx)

		lot, err := svcLot.Lot(act.LotID(), false)
		if err != nil {
			logger.Error("get lot failed", zap.Error(err))
			return err
		}

		if lot.BookedAt == nil {
			logger.Warn("lot not booked")
			return lotNotBooked
		}

		bcksvc := service.NewBackService(auc.cfgBackService, act.Executor(), lot)

		ld, err := bcksvc.GetBookingWindow()
		if err != nil {
			logger.Error("get booking window failed", zap.Error(err))
			return err
		}

		act.SetLoad(ld)

		return nil
	})
}

func (auc *auction) UpdateLotAct(cmd interfaces.EditActCommander) error {
	logger := auc.logger.Named("edit_object_act")
	var lot *core.Lot
	var err error

	return auc.tx(func(tx *db.Tx) error {
		lotSvc := service.NewLotService(tx)
		lot, err = lotSvc.Lot(cmd.LotID(), false)
		if err != nil {
			logger.Error("get lot for updating cmd failed", zap.Error(err))
			return err
		}

		complete, err := auc.ejectComplete(lot, logger)
		if err != nil {
			return err
		}

		// получаем флаг разрешающий редактировать.
		allowChange := complete.AllowChange
		if allowChange == nil || *allowChange == 0 {
			logger.Warn("secondary attempt to change the act")
			return errors.BadRequest("you can not secondary change the act")
		}
		// запрещаем дальнейшее редактирование
		i := 0
		complete.AllowChange = &i
		act := cmd.GetObjectAct()
		complete.ActDate = act.Date.UTC()
		complete.ActNumber = act.ActNumber

		if complete.Validate() != nil {
			logger.Error("validate new data for lot.complete failed", zap.Error(err))
			return err
		}
		// преобразовываем обратно в json
		b, err := json.Marshal(complete)
		if err != nil {
			logger.Error("new complete data marshalling failed", zap.Error(err))
			return err
		}
		lot.Complete = object.JSONData(b)

		// сохраняем изменения по акту
		if err := lotSvc.SaveLot(lot); err != nil {
			logger.Error("update lot after changed act data failed", zap.Error(err))
			return err
		}

		historyService := service.NewHistoryService(tx)
		if err := historyService.TheActOfTheLotIsChanged(cmd.Executor().ID, lot); err != nil {
			logger.Error("save history of act change failed", zap.Error(err))
			return err
		}

		// отправляем менеджеру запрос, для синхронизации данных
		act.ObjectID = lot.ObjectID
		backSvc := service.NewBackService(auc.cfgBackService, cmd.Executor(), lot)
		if err := backSvc.ChangeActData(act); err != nil {
			logger.Error("send new complete data to manager failed", zap.Error(err))
			return err
		}

		return nil
	})
}

func (auc *auction) AutoBookingLot(cmd interfaces.AutoBookingCommander) error {
	lot := cmd.GetLot()

	now := time.Now().UTC()
	lot.BookedAt = &now
	return auc.CreateLot(cmd)
}
// Открытие доступа на редактирование акта
func (auc *auction) AllowChangeAct(cmd interfaces.AllowChangeActCommander) error {
	logger := auc.logger.Named("allow_act_editing")

	return auc.tx(func(tx *db.Tx) error {
		lotSvc := service.NewLotService(tx)
		lot, err := lotSvc.Lot(cmd.LotID(), false)
		if err != nil {
			logger.Error("get lot by id failed", zap.Error(err))
			return err
		}

		complete, err := auc.ejectComplete(lot, logger)
		if err != nil {
			if err == actNotFound {
				// TODO: Временный костыль,  пока не поправят логистику.
				// для Агро и Сельта нет актов.
				betsService := service.NewBetService(tx)
				bet, err := betsService.GetWinnerBetByLotId(lot)
				if err != nil {
					return betNotFound
				}

				exceptions := []uint{
					// todo: надеюсь это уберут когда нибудь
					8, // selta
					140, // selta
					145, // agro
				}

				for _, userID := range exceptions {
					if userID == bet.UserID {
						return nil
					}
				}

				return actNotFound
			}
			return err
		}

		allowChange := cmd.GetAllowChange()
		if complete.AllowChange == &allowChange {
			return nil
		}

		complete.AllowChange = &allowChange

		// преобразовываем обратно в json
		b, err := json.Marshal(complete)
		if err != nil {
			logger.Error("new complete data marshalling failed", zap.Error(err))
			return err
		}
		lot.Complete = object.JSONData(b)

		// сохраняем изменения по акту
		if err := lotSvc.SaveLot(lot); err != nil {
			logger.Error("update lot after changed act data failed", zap.Error(err))
			return err
		}

		// оставляем след в истории
		historyService := service.NewHistoryService(tx)
		if err := historyService.AllowChangeForAct(cmd.Executor().ID, lot, cmd); err != nil {
			logger.Error("save history of act change failed", zap.Error(err))
			return err
		}
		return nil
	})
}

func (auc *auction) ejectComplete(lot *core.Lot, logger *zap.Logger) (*json_fileds.Complete, error) {
	// изменять акт можно только по завершенным лотам
	if lot.CompletedAt == nil {
		logger.Error("an attempt to change the number and date of the act for the flight unfinished")
		return nil, errors.BadRequest("you can not change the data on the unfinished lot")
	}

	// лот завершен, но акт еще не был создан
	if lot.Complete == nil {
		logger.Error("act not founded")
		return nil, actNotFound
	}

	// преобразовываем наш  json в структуру
	complete := &json_fileds.Complete{}
	if err := json.Unmarshal(lot.Complete, complete); err != nil {
		logger.Error(fmt.Sprintf("unmarshal lot.complete data failed, lot_id = %d", lot.ID), zap.Error(err))
		return nil, err
	}
	return complete, nil
}

func (auc *auction) tx(handler func(tx *db.Tx) error) error {
	logger := auc.logger.Named("tx")

	tx, err := auc.db.Begin()
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

func syncLots(prcs map[uint]process.Process, lots ...*core.Lot) {
	for i := 0; i < len(lots); i++ {
		lot := lots[i]
		if prc, ok := prcs[lot.ID]; ok {
			prc.Sync(lot)
		}
	}
}
