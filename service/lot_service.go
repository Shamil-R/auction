package service

import (
	"database/sql"
	"fmt"
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/db"
	"gitlab/nefco/auction/helper"
)

type lotService struct {
	tx *db.Tx
}

func NewLotService(tx *db.Tx) *lotService {
	return &lotService{tx}
}

func (svc *lotService) Lots(filter core.LotsFilter) ([]*core.Lot, error) {
	lots := []*core.Lot{}

	arg := map[string]interface{}{}

	var joinBets, whereGroups, whereState, orderLots string

	if filter != nil {
		whereGroups = `
			AND lots.group_key IN (:groups)
		`
		arg["groups"] = filter.Groups()
		if len(filter.Group()) > 0 {
			exists, _ := helper.InArray(filter.Group(), filter.Groups())
			if !exists {
				return nil, nil
			}
			arg["groups"] = filter.Group()
		}

		switch filter.State() {
		case core.FilterStateActive:
			if filter.Executor().Blocked {
				return lots, nil
			}
			whereState = `
				AND lots.booked_at IS NULL
			`
		case core.FilterStateBooked:
			whereState = `
				AND lots.booked_at IS NOT NULL 
				AND lots.completed_at IS NULL
			`
		case core.FilterStateCompleted:
			whereState = `
				AND lots.completed_at IS NOT NULL
			`
			orderLots = `
				ORDER BY completed_at desc
			`
		case core.FilterStateNoCompleted:
			whereState = `
				AND lots.completed_at IS NULL
			`
		}

		if filter.State() == core.FilterStateBooked ||
			filter.State() == core.FilterStateCompleted ||
			filter.State() == core.FilterStateNoCompleted {
			joinBets = `
				JOIN bets ON bets.lot_id = lots.id 
					AND bets.user_id = :user_id
					AND bets.winner = 1
					AND bets.deleted_at IS NULL
			`
			arg["user_id"] = filter.Executor().ID
		}
	} else {
		whereState = `
			AND lots.manual_booked = 0
			AND lots.confirmed_at IS NULL
		`
	}

	query := fmt.Sprintf(`
		SELECT
			lots.id, 
			lots.rules,
			lots.created_at,
			lots.updated_at,
			lots.booked_at,
			lots.manual_booked,
			lots.confirmed_at,
			lots.completed_at,
			lots.object_id,
			lots.object,
			lots.confirm,
			lots.group_key,
			lots.user_id,
			lots.urgent
		FROM lots
		%s
		WHERE lots.deleted_at IS NULL
		%s
		%s
		%s
	`, joinBets, whereGroups, whereState, orderLots)

	if err := svc.tx.Select(&lots, query, arg); err != nil {
		return nil, err
	}

	if len(lots) == 0 {
		return lots, nil
	}

	idsBet := make([]uint, len(lots))
	idsUser := make([]uint, len(lots))

	for i, lot := range lots {
		idsBet[i] = lot.ID
		idsUser[i] = lot.UserID
	}

	/*arg = map[string]interface{}{
		"ids": idsBet,
	}*/

	query = `
		SELECT id, value, winner, created_at, lot_id, user_id
		FROM bets
		WHERE lot_id IN (:ids) AND deleted_at IS NULL
	`

	bets := []*core.Bet{}

	// todo: test version with many argument
	argLen := len(idsBet)
	delimiter := 2000
	argParts := int(argLen/delimiter) + 1

	for i := 0; i < argParts; i++ {
		betPart := []*core.Bet{}

		leftBorder := i * delimiter
		var rightBorder int
		if argLen < (i+1)*delimiter {
			rightBorder = argLen
		} else {
			rightBorder = (i + 1) * delimiter
		}

		arg = map[string]interface{}{
			"ids": idsBet[leftBorder:rightBorder],
		}

		if err := svc.tx.Select(&betPart, query, arg); err != nil {
			return nil, err
		}
		bets = append(bets, betPart...)
	}

	tempLots := make([]*core.Lot, 0, len(lots))

	for _, lot := range lots {
		rest := make([]*core.Bet, 0, len(bets))
		for _, bet := range bets {
			if lot.ID == bet.LotID {
				lot.AddBet(bet)
			} else {
				rest = append(rest, bet)
			}
		}
		if filter != nil && filter.OnlyWithBets() {
			if len(lot.Bets) > 0 {
				tempLots = append(tempLots, lot)
			}
		} else {
			tempLots = append(tempLots, lot)
		}
		bets = rest
	}

	lots = tempLots

	/*arg = map[string]interface{}{
		"ids": idsUser,
	}*/

	query = `
		SELECT *
		FROM users
		WHERE id IN (:ids)
	`

	users := []*core.User{}

	// todo: test version with many argument
	argLen = len(idsUser)
	delimiter = 2000
	argParts = int(argLen/delimiter) + 1

	for i := 0; i < argParts; i++ {
		userPart := []*core.User{}

		leftBorder := i * delimiter
		var rightBorder int
		if argLen < (i+1)*delimiter {
			rightBorder = argLen
		} else {
			rightBorder = (i + 1) * delimiter
		}

		arg = map[string]interface{}{
			"ids": idsUser[leftBorder:rightBorder],
		}

		if err := svc.tx.Select(&userPart, query, arg); err != nil {
			return nil, err
		}
		users = append(users, userPart...)
	}

	for _, lot := range lots {
		for _, user := range users {
			if lot.UserID == user.ID {
				lot.User = user
				break
			}
		}
	}

	return lots, nil
}

func (svc *lotService) Lot(id uint, trashed bool) (*core.Lot, error) {
	lot := &core.Lot{}

	arg := map[string]interface{}{
		"id": id,
	}

	var where string
	if !trashed {
		where = "AND deleted_at IS NULL"
	}

	query := fmt.Sprintf(`
		SELECT * FROM lots 
		WHERE id = :id
		%s
	`, where)

	if err := svc.tx.Get(lot, query, arg); err != nil {
		if err == sql.ErrNoRows {
			return nil, core.LotNotFound
		}
		return nil, err
	}

	arg = map[string]interface{}{
		"id": lot.UserID,
	}

	query = `
		SELECT * FROM users 
		WHERE id = :id
	`

	user := &core.User{}

	if err := svc.tx.Get(user, query, arg); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	lot.User = user

	arg = map[string]interface{}{
		"id": lot.ID,
	}

	query = `
		SELECT id, value, winner, created_at, lot_id, user_id
		FROM bets
		WHERE lot_id = :id AND deleted_at IS NULL
	`

	bets := []*core.Bet{}

	if err := svc.tx.Select(&bets, query, arg); err != nil {
		return nil, err
	}

	for _, bet := range bets {
		if lot.ID == bet.LotID {
			lot.AddBet(bet)
		}
	}

	return lot, nil
}

func (svc *lotService) LotByObject(groupKey string, objectID uint) (*core.Lot, error) {
	lot := &core.Lot{}

	arg := map[string]interface{}{
		"group_key": groupKey,
		"object_id": objectID,
	}

	query := `
		SELECT * FROM lots 
		WHERE group_key = :group_key
		AND object_id = :object_id
	`

	if err := svc.tx.Get(lot, query, arg); err != nil {
		if err == sql.ErrNoRows {
			return nil, core.LotNotFound
		}
		return nil, err
	}

	return lot, nil
}

func (svc *lotService) CreateLot(lot *core.Lot) error {
	lot.CreatedAt = now()
	lot.UpdatedAt = now()

	query := `
		INSERT INTO lots (
			rules, 
			created_at,
			updated_at,
			booked_at,
			manual_booked,
			object_id,
			object,
			group_key,
			user_id,
			urgent
		)
		VALUES (
			:rules, 
			:created_at,
			:updated_at,
			:booked_at,
			:manual_booked,
			:object_id,
			:object,
			:group_key,
			:user_id,
			:urgent
		)
	`

	result, err := svc.tx.Exec(query, lot)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	lot.ID = uint(id)

	return nil
}

func (svc *lotService) SaveLot(lot *core.Lot) error {
	lot.UpdatedAt = now()

	query := `
		UPDATE lots SET 
			rules = :rules, 
			updated_at = :updated_at,
			deleted_at = :deleted_at,
			booked_at = :booked_at,
			manual_booked = :manual_booked,
			confirmed_at = :confirmed_at,
			completed_at = :completed_at,
			object = :object,
			confirm = :confirm,
			complete = :complete,
			urgent = :urgent
		WHERE id = :id`

	if _, err := svc.tx.Exec(query, lot); err != nil {
		return err
	}

	return nil
}

func (svc *lotService) DeleteLot(lot *core.Lot) error {
	n := now()

	lot.DeletedAt = &n

	query := `UPDATE lots SET deleted_at = :deleted_at WHERE id = :id`

	if _, err := svc.tx.Exec(query, lot); err != nil {
		return err
	}

	return nil
}
