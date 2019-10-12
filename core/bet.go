package core

import "time"

type Bet struct {
	ID        uint64     `json:"id" db:"id"`
	Value     uint       `json:"value" db:"value" validate:"required,gt=0"`
	Winner    bool       `json:"winner" db:"winner"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
	LotID     uint       `json:"-" db:"lot_id"`
	UserID    uint       `json:"user_id" db:"user_id" validate:"required"`
}

type BetService interface {
	MinBet(lotID uint) (*Bet, error)
	Bets(lotID uint) ([]*Bet, error)
	Bet(betID uint) (*Bet, error)
	CreateBet(bet *Bet) error
	SaveBet(bet *Bet) error
	DeleteBet(bet *Bet) error
	ClearBets(lotID uint) error
	ClearBetsBefore(lotID uint, t time.Time) error
}
