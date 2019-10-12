package command

import (
	"gitlab/nefco/auction/core"
	"net/http"
)

const CommandGetLot = "get.lot"

type GetLot struct {
	*lot
	l *core.Lot
}

func newGetLot(user *core.User) *GetLot {
	return &GetLot{
		lot: newLot(CommandGetLot, AccessManagerUser, user),
	}
}

func (cmd *GetLot) SetLot(lot *core.Lot) {
	cmd.l = lot
}

func (cmd *GetLot) Event() Event {
	return &struct {
		*event
		*core.Lot
	}{
		newSucces(cmd.name, http.StatusOK),
		cmd.l,
	}
}
