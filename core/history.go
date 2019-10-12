package core

import (
	"time"
)

const (
	HistoryLotAdded              = "lot_added"
	HistoryLotUpdated            = "lot_updated"
	HistoryLotDeleted            = "lot_deleted"
	HistoryLotBooked             = "lot_booked"
	HistoryLotCompleted          = "lot_completed"
	HistoryLotNoWinner           = "lot_no_winner"
	HistoryBetPlaced             = "bet_placed"
	HistoryBetCanceled           = "bet_canceled"
	HistoryLotNoConfirm          = "lot_no_confirm"
	HistoryLotConfirmed          = "lot_confirmed"
	HistoryLotConfirmEdited      = "lot_confirm_edited"
	HistoryLotDeleteConfirmation = "lot_delete_confirmation"
	HistoryLotBetAccept          = "lot_bet_accept"
	HistoryActChanged			 = "act_changed"
	HistoryAllowChangedForAct    = "act_allow_changed=%d"
)

type History struct {
	ID                 uint      `json:"id" db:"id"`
	Action             string    `json:"action" db:"action"`
	Rule               *string   `json:"rule,omitempty" db:"rule"`
	RulePrice          *uint     `json:"rule_price,omitempty" db:"rule_price"`
	CurrentPrice       *uint     `json:"current_price,omitempty" db:"current_price"`
	CurrentPriceUserID *uint     `json:"current_price_user_id,omitempty" db:"current_price_user_id"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	LotID              uint      `json:"lot_id" db:"lot_id"`
	UserID             uint      `json:"user_id" db:"user_id"`
}

type HistoryService interface {
	LotAdded(userID uint, lot *Lot) error
	LotUpdated(userID uint, lot *Lot) error
	LotDeleted(userID uint, lot *Lot) error
	LotBooked(userID uint, lot *Lot) error
	LotCompleted(userID uint, lot *Lot) error
	LotNotWinner(userID uint, lot *Lot) error
	BetPlaced(userID uint, lot *Lot) error
	BetCanceled(userID uint, lot *Lot) error
	LotConfirmed(userID uint, lot *Lot) error
	LotConfirmEdited(userID uint, lot *Lot) error
	LotNoConfirm(userID uint, lot *Lot) error
	LotDeleteConfirmation(userID uint, lot *Lot) error
	LotBetAccept(userID uint, lot *Lot) error
	History(lotID uint) ([]*History, error)
}
