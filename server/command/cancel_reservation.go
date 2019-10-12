package command

import "gitlab/nefco/auction/core"

const CommandCancelReservation = "cancel.reservation"

type CancelReservation struct {
	*lot
}

func newCancelReservation(user *core.User) *CancelReservation {
	return &CancelReservation{
		lot: newLot(CommandCancelReservation, AccessUser, user),
	}
}
