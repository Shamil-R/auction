package command

import (
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/core/object"

	"github.com/labstack/echo"
)

const CommandConfirmLot = "confirm.lot"

type ConfirmLot struct {
	*lot
	l           *core.Lot
	ConfirmInfo object.JSONData `json:"info"`
}

func newConfirmLot(user *core.User) *ConfirmLot {
	return &ConfirmLot{
		lot: newLot(CommandConfirmLot, AccessUser, user),
	}
}

func (cmd *ConfirmLot) Info() object.JSONData {
	return cmd.ConfirmInfo
}

func (cmd *ConfirmLot) GetLot() *core.Lot {
	return cmd.l
}

func (cmd *ConfirmLot) SetLot(lot *core.Lot) {
	cmd.l = lot
}

func (cmd *ConfirmLot) eject(ctx echo.Context) error {
	return ctx.Bind(cmd)
}
