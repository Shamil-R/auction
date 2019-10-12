package command

import (
	"gitlab/nefco/auction/core"
	"net/http"
)

const CommandGetBookingWindow = "get.booking_window"

type GetBookingWindow struct {
	*lot
	ld []core.LoadDate
}

func newGetBookingWindow(user *core.User) *GetBookingWindow {
	return &GetBookingWindow{
		lot: newLot(CommandGetBookingWindow, AccessUser, user),
	}
}

func (cmd *GetBookingWindow) SetLoad(ld []core.LoadDate) {
	cmd.ld = ld
}

func (cmd *GetBookingWindow) Event() Event {
	e := eventGetBookingWindow(cmd.ld)
	return &e
}

type eventGetBookingWindow []core.LoadDate

func (evt *eventGetBookingWindow) Event() string {
	return successName(CommandGetBookingWindow)
}

func (evt *eventGetBookingWindow) Code() int {
	return http.StatusOK
}
