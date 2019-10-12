package service

import (
	"fmt"
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/db"
	"gitlab/nefco/auction/interfaces"
)

type historyService struct {
	tx *db.Tx
}

func NewHistoryService(tx *db.Tx) *historyService {
	return &historyService{tx}
}

func (svc *historyService) LotAdded(userID uint, lot *core.Lot) error {
	return svc.create(core.HistoryLotAdded, userID, lot)
}

func (svc *historyService) LotUpdated(userID uint, lot *core.Lot) error {
	return svc.create(core.HistoryLotUpdated, userID, lot)
}

func (svc *historyService) LotDeleted(userID uint, lot *core.Lot) error {
	return svc.create(core.HistoryLotDeleted, userID, lot)
}

func (svc *historyService) LotBooked(userID uint, lot *core.Lot) error {
	return svc.create(core.HistoryLotBooked, core.RootUserID, lot)
}

func (svc *historyService) LotCompleted(userID uint, lot *core.Lot) error {
	return svc.create(core.HistoryLotCompleted, core.RootUserID, lot)
}

func (svc *historyService) LotNotWinner(userID uint, lot *core.Lot) error {
	return svc.create(core.HistoryLotNoWinner, core.RootUserID, lot)
}

func (svc *historyService) BetPlaced(userID uint, lot *core.Lot) error {
	return svc.create(core.HistoryBetPlaced, userID, lot)
}

func (svc *historyService) BetCanceled(userID uint, lot *core.Lot) error {
	return svc.create(core.HistoryBetCanceled, userID, lot)
}

func (svc *historyService) LotNoConfirm(userID uint, lot *core.Lot) error {
	return svc.create(core.HistoryLotNoConfirm, core.RootUserID, lot)
}

func (svc *historyService) LotConfirmed(userID uint, lot *core.Lot) error {
	return svc.create(core.HistoryLotConfirmed, userID, lot)
}

func (svc *historyService) LotConfirmEdited(userID uint, lot *core.Lot) error {
	return svc.create(core.HistoryLotConfirmEdited, userID, lot)
}

func (svc *historyService) LotDeleteConfirmation(userID uint, lot *core.Lot) error {
	return svc.create(core.HistoryLotDeleteConfirmation, userID, lot)
}

func (svc *historyService) LotBetAccept(userID uint, lot *core.Lot) error {
	return svc.create(core.HistoryLotBetAccept, userID, lot)
}

func (svc *historyService) TheActOfTheLotIsChanged(userID uint, lot *core.Lot) error {
	return svc.create(core.HistoryActChanged, userID, lot)
}

func (svc *historyService) AllowChangeForAct(userID uint, lot *core.Lot, getChange interfaces.ActAllowChangeParamGetter) error {
	return svc.create(fmt.Sprintf(core.HistoryAllowChangedForAct, getChange.GetAllowChange()), userID, lot)
}

func (svc *historyService) History(lotID uint) ([]*core.History, error) {
	histories := []*core.History{}

	arg := map[string]interface{}{
		"lot_id": lotID,
	}

	query := `SELECT * FROM history WHERE lot_id = :lot_id`

	if err := svc.tx.Select(&histories, query, arg); err != nil {
		return nil, err
	}

	return histories, nil
}

func (svc *historyService) create(action string, userID uint, lot *core.Lot) error {
	var currentUserID uint

	if lot.Bet != nil {
		currentUserID = lot.Bet.UserID
	}

	history := &core.History{
		Action:             action,
		CreatedAt:          now(),
		LotID:              lot.ID,
		UserID:             userID,
		Rule:               &lot.Rule,
		RulePrice:          &lot.RulePrice,
		CurrentPrice:       &lot.CurrentPrice,
		CurrentPriceUserID: &currentUserID,
	}

	query := `
		INSERT INTO history (
			action, 
			[rule],
			rule_price,
			current_price,
			current_price_user_id,
			created_at,
			lot_id, 
			user_id
		)
		VALUES (
			:action, 
			:rule,
			:rule_price,
			:current_price,
			:current_price_user_id,
			:created_at,
			:lot_id, 
			:user_id
		)
	`

	if _, err := svc.tx.Exec(query, history); err != nil {
		return err
	}

	return nil
}
