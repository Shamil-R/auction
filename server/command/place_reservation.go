package command

import "gitlab/nefco/auction/core"

const CommandPlaceReservation = "place.reservation"

type PlaceReservation struct {
	*lot
}

func newPlaceReservation(user *core.User) *PlaceReservation {
	return &PlaceReservation{
		lot: newLot(CommandPlaceReservation, AccessUser, user),
	}
}
