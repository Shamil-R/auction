package service

import (
	"database/sql"
	"fmt"
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/db"
	"gitlab/nefco/auction/errors"
	"time"
)

var (
	BetNotFound = errors.NotFound("Bet not found")
)

type betService struct {
	tx *db.Tx
}

func NewBetService(tx *db.Tx) *betService {
	return &betService{tx}
}

func (svc *betService) MinBet(lotID uint) (*core.Bet, error) {
	var bet *core.Bet

	arg := map[string]interface{}{
		"lot_id": lotID,
	}

	query := `
		SELECT 
			id,
			value,
			winner,
			created_at,
			deleted_at,
			lot_id,
			user_id
		FROM bets 
		WHERE lot_id = :lot_id 
		AND deleted_at IS NULL
	`

	bets := []*core.Bet{}

	if err := svc.tx.Select(&bets, query, arg); err != nil {
		return nil, err
	}

	for _, b := range bets {
		if bet == nil || bet.Value > b.Value {
			bet = b
		}
	}

	return bet, nil
}

func (svc *betService) Bets(lotID uint) ([]*core.Bet, error) {
	bets := []*core.Bet{}

	arg := map[string]interface{}{
		"lot_id": lotID,
	}

	query := `
		SELECT 
			id,
			value,
			winner,
			created_at,
			deleted_at,
			lot_id,
			user_id
		FROM bets 
		WHERE lot_id = :lot_id 
		AND deleted_at IS NULL
		ORDER BY value DESC
	`

	if err := svc.tx.Select(&bets, query, arg); err != nil {
		return nil, err
	}

	return bets, nil
}

func (svc *betService) Bet(betID uint) (*core.Bet, error) {
	bet := &core.Bet{}

	arg := map[string]interface{}{
		"id": betID,
	}

	query := fmt.Sprintf(`
		SELECT * FROM bets 
		WHERE id = :id
		AND deleted_at IS NULL
	`)

	if err := svc.tx.Get(bet, query, arg); err != nil {
		if err == sql.ErrNoRows {
			return nil, BetNotFound
		}
		return nil, err
	}

	return bet, nil
}

func (svc *betService) CreateBet(bet *core.Bet) error {
	bet.CreatedAt = now()

	query := `
		INSERT INTO bets (
			value, 
			winner,
			created_at,
			lot_id, 
			user_id
		)
		VALUES (
			:value,
			:winner, 
			:created_at,
			:lot_id, 
			:user_id
		)
	`

	result, err := svc.tx.Exec(query, bet)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	bet.ID = uint64(id)

	return nil
}

func (svc *betService) SaveBet(bet *core.Bet) error {
	query := `UPDATE bets SET winner = :winner WHERE id = :id`

	_, err := svc.tx.Exec(query, bet)
	if err != nil {
		return err
	}

	return nil
}

func (svc *betService) DeleteBet(bet *core.Bet) error {
	n := now()

	bet.DeletedAt = &n

	query := `UPDATE bets SET deleted_at = :deleted_at WHERE id = :id`

	_, err := svc.tx.Exec(query, bet)
	if err != nil {
		return err
	}

	return nil
}

func (svc *betService) ClearBets(lotID uint) error {
	arg := struct {
		DeletedAt time.Time `db:"deleted_at"`
		LotID     uint      `db:"lot_id"`
	}{
		now(),
		lotID,
	}

	query := `
		UPDATE bets SET deleted_at = :deleted_at
		WHERE lot_id = :lot_id
		AND deleted_at IS NULL
	`

	if _, err := svc.tx.Exec(query, arg); err != nil {
		return err
	}

	return nil
}

func (svc *betService) ClearBetsBefore(lotID uint, t time.Time) error {
	arg := struct {
		DeletedAt time.Time `db:"deleted_at"`
		LotID     uint      `db:"lot_id"`
		CreatedAt time.Time `db:"created_at"`
	}{
		now(),
		lotID,
		t,
	}

	query := `
		UPDATE bets SET deleted_at = :deleted_at
		WHERE lot_id = :lot_id
		AND deleted_at IS NULL
		AND created_at < :created_at
	`

	if _, err := svc.tx.Exec(query, arg); err != nil {
		return err
	}

	return nil
}

func (svc *betService) GetWinnerBetByLotId(lot *core.Lot) (*core.Bet, error) {
	query := `
		SELECT 
			value, 
			winner,
			created_at,
			lot_id, 
			user_id
		FROM bets 
		where lot_id = :lot_id and winner = 1
	`
	bet := &core.Bet{}

	arg := map[string]interface{}{
		"lot_id": lot.ID,
	}

	if err := svc.tx.Get(bet, query, arg); err != nil {
		if err == sql.ErrNoRows {
			return nil, BetNotFound
		}
		return nil, err
	}
	return bet, nil
}