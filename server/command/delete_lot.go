package command

import (
	"gitlab/nefco/auction/core"
)

const CommandDeleteLot = "delete.lot"

type DeleteLot struct {
	*lot
}

func newDeleteLot(user *core.User) *DeleteLot {
	return &DeleteLot{
		lot: newLot(CommandDeleteLot, AccessManager, user),
	}
}
