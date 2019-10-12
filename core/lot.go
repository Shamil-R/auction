package core

import (
	"database/sql/driver"
	"encoding/json"
	"gitlab/nefco/auction/core/object"
	"gitlab/nefco/auction/db"
	"gitlab/nefco/auction/errors"
	"time"

	validator "gopkg.in/go-playground/validator.v9"
)

var (
	LotNotFound = errors.NotFound("Lot not found")
)

type RuleConfig struct {
	Type     string          `json:"type" validate:"required,rule"`
	Start    string          `json:"start" validate:"required"` // TODO: добавить валидацию значения, должно передаваться время
	Duration Duration        `json:"duration" validate:"required"`
	Props    json.RawMessage `json:"props" validate:"required"`
}

type Rules []*RuleConfig

func (r Rules) Value() (driver.Value, error) {
	return db.JSONValue(&r)
}

func (r *Rules) Scan(src interface{}) error {
	return db.JSONScan(src, r)
}

type Lot struct {
	ID           uint            `json:"id" db:"id"`
	Rules        Rules           `json:"rules" db:"rules" validate:"required,min=1,dive,required" groups:"manager"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time      `json:"-" db:"deleted_at"`
	BookedAt     *time.Time      `json:"booked_at,omitempty" db:"booked_at"`
	ManualBooked bool            `json:"manual_booked" db:"manual_booked"`
	ConfirmedAt  *time.Time      `json:"confirmed_at,omitempty" db:"confirmed_at"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty" db:"completed_at"`
	GroupKey     string          `json:"group_key" db:"group_key" validate:"required"`
	ObjectID     uint            `json:"object_id" db:"object_id" validate:"required"`
	Object       object.Object   `json:"object" db:"object" validate:"required"`
	Confirm      object.JSONData `json:"confirm,omitempty" db:"confirm"`
	Complete     object.JSONData `json:"complete,omitempty" db:"complete"`
	UserID       uint            `json:"-" db:"user_id"`
	User         *User           `json:"-"`
	Bet          *Bet            `json:"bet,omitempty"`
	BetStep      uint            `json:"bet_step"`
	BasePrice    uint            `json:"base_price,omitempty"`
	CurrentPrice uint            `json:"current_price,omitempty"`
	UserPrice    uint            `json:"user_price,omitempty"`
	Rule         string          `json:"rule,omitempty"`
	RulePrice    uint            `json:"-"`
	End          time.Time       `json:"end,omitempty"`
	Rest         uint            `json:"rest"`
	Bets         []*Bet          `json:"bets" groups:"manager"`
	Urgent       bool            `json:"urgent" db:"urgent"`
}

func (l *Lot) AddBet(bet *Bet) {
	l.Bets = append(l.Bets, bet)
}

func (l *Lot) RemoveBet(bet *Bet) {
	bets := make([]*Bet, 0, len(l.Bets))
	for _, b := range l.Bets {
		if bet.ID != b.ID {
			bets = append(bets, b)
		}
	}
	l.Bets = bets
}

func (l *Lot) CurrentBet() *Bet {
	var bet *Bet
	for _, b := range l.Bets {
		if b.Winner {
			return b
		}
		if bet == nil || bet.Value > b.Value {
			bet = b
		}
	}
	return bet
}

func (l *Lot) BetByID(betID uint64) *Bet {
	for _, b := range l.Bets {
		if b.ID == betID {
			return b
		}
	}
	return nil
}

func (l *Lot) UserBet(userID uint) *Bet {
	var bet *Bet
	for _, b := range l.Bets {
		if b.UserID == userID && (bet == nil || bet.Value > b.Value) {
			bet = b
		}
	}
	return bet
}

func (l *Lot) ClearBets() {
	l.Bets = []*Bet{}
}

func (l *Lot) ClearBetsBefore(t time.Time) {
	bets := make([]*Bet, 0, len(l.Bets))
	for _, b := range l.Bets {
		if b.CreatedAt.After(t) {
			bets = append(bets, b)
		}
	}
	l.Bets = bets
}

func (l *Lot) UpdatePrice(userID uint) {
	curBet := l.CurrentBet()
	if curBet != nil {
		l.CurrentPrice = curBet.Value
		l.Bet = curBet
	}
	userBet := l.UserBet(userID)
	if userBet != nil {
		l.UserPrice = userBet.Value
	} else {
		l.UserPrice = 0
	}
}

const (
	FilterStateNone        = ""
	FilterStateAll         = "all"
	FilterStateActive      = "active"
	FilterStateBooked      = "booked"
	FilterStateCompleted   = "completed"
	FilterStateNoCompleted = "no_completed"
)

type GroupsFilter interface {
	Groups() []string
}

type LotsFilter interface {
	GroupsFilter
	Executor() *User
	Group() string
	State() string
	LotID() uint
	ObjectID() uint
	StartPrice() uint
	EndPrice() uint
	OnlyWithBets() bool
	Filter() object.ObjectFilter
}

type LotService interface {
	Lots(filter LotsFilter) ([]*Lot, error)
	Lot(id uint, trashed bool) (*Lot, error)
	LotByObject(groupKey string, objectID uint) (*Lot, error)
	CreateLot(lot *Lot) error
	SaveLot(lot *Lot) error
	DeleteLot(lot *Lot) error
}

func ValidateState(validate *validator.Validate) {
	validate.RegisterValidation("state",
		func(fl validator.FieldLevel) bool {
			state := fl.Field().String()
			res := state == FilterStateNone ||
				state == FilterStateAll ||
				state == FilterStateActive ||
				state == FilterStateBooked ||
				state == FilterStateCompleted ||
				state == FilterStateNoCompleted
			return res
		},
	)
}
