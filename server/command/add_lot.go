package command

import (
	"gitlab/nefco/auction/core"
	"net/http"

	"github.com/labstack/echo"
)

const CommandAddLot = "add.lot"

type AddLot struct {
	*base
	*core.Lot
}

func newAddLot(user *core.User) *AddLot {
	return &AddLot{
		base: newBase(CommandAddLot, AccessManager, user),
		Lot:  &core.Lot{},
	}
}

func (cmd *AddLot) GetLot() *core.Lot {
	return cmd.Lot
}

func (cmd *AddLot) Event() Event {
	return &struct {
		*event
		*core.Lot
	}{
		newSucces(cmd.name, http.StatusCreated),
		cmd.Lot,
	}
}

func (cmd *AddLot) eject(ctx echo.Context) error {
	return ctx.Bind(cmd)
}
