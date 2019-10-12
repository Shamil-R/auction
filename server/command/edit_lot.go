package command

import (
	"gitlab/nefco/auction/core"
	"net/http"

	"github.com/labstack/echo"
)

const CommandEditLot = "edit.lot"

type EditLot struct {
	*lot
	*core.Lot
}

func newEditLot(user *core.User) *EditLot {
	return &EditLot{
		lot: newLot(CommandEditLot, AccessManager, user),
		Lot: &core.Lot{},
	}
}

func (cmd *EditLot) GetLot() *core.Lot {
	return cmd.Lot
}

func (cmd *EditLot) SetLot(lot *core.Lot) {
	cmd.Lot = lot
}

func (cmd *EditLot) Event() Event {
	return &struct {
		*event
		*core.Lot
	}{
		newSucces(cmd.name, http.StatusOK),
		cmd.Lot,
	}
}

func (cmd *EditLot) eject(ctx echo.Context) error {
	return ctx.Bind(cmd)
}
