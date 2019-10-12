package command

import (
	"github.com/labstack/echo"
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/errors"
	"net/http"
)

const CommandAutoBooking = "autoBooking"


type AutoBooking struct {
	*base
	*core.Lot
}

func newAutoBooking(user *core.User) *AutoBooking {
	return &AutoBooking{
		base: newBase(CommandAutoBooking, AccessManager, user),
		Lot: &core.Lot{},
	}
}

func (cmd *AutoBooking) GetLot() *core.Lot {
	return cmd.Lot
}

func (cmd *AutoBooking) GetUserId() uint {
	return cmd.Lot.Bet.UserID
}

func (cmd *AutoBooking) eject(ctx echo.Context) error {
	if err := ctx.Bind(cmd); err != nil {
		return err
	}
	bet := cmd.Lot.Bet

	if bet == nil {
		return errors.BadRequest("object lot.Bet is not be empty")
	}
	if bet.UserID == 0 {
		return errors.BadRequest("field user_id for bet is not be empty")
	}
	if bet.Value == 0 {
		return errors.BadRequest("field value is not be empty")
	}
	return nil
}

func (cmd *AutoBooking) Event() Event {
	return &struct {
		*event
		*core.Lot
	}{
		newSucces(cmd.name, http.StatusCreated),
		cmd.Lot,
	}
}
