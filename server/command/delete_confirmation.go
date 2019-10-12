package command

import "gitlab/nefco/auction/core"

const CommandDeleteConfirmation = "delete.confirmation"

type DeleteConfirmation struct {
	*lot
}

func newDeleteConfirmation(user *core.User) *DeleteConfirmation {
	return &DeleteConfirmation{
		lot: newLot(CommandDeleteConfirmation, AccessUser, user),
	}
}
