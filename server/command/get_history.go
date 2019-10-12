package command

import (
	"gitlab/nefco/auction/core"
	"net/http"
)

const CommandGetHistory = "get.history"

type GetHistory struct {
	*lot
	histories []*core.History
}

func newGetHistory(user *core.User) *GetHistory {
	return &GetHistory{
		lot: newLot(CommandGetHistory, AccessManager, user),
	}
}

func (cmd *GetHistory) SetHistory(histories []*core.History) {
	cmd.histories = histories
}

func (cmd *GetHistory) Event() Event {
	e := eventGetHistory(cmd.histories)
	return &e
}

type eventGetHistory []*core.History

func (evt *eventGetHistory) Event() string {
	return successName(CommandGetHistory)
}

func (evt *eventGetHistory) Code() int {
	return http.StatusOK
}
