package command

import (
	"gitlab/nefco/auction/core"
	"net/http"
)

const CommandGetUsers = "get.users"

type GetUsers struct {
	*base
	users []*core.User
}

func newGetUsers(user *core.User) *GetUsers {
	return &GetUsers{
		base: newBase(CommandGetUsers, AccessRoot, user),
	}
}

func (cmd *GetUsers) SetUsers(users []*core.User) {
	cmd.users = users
}

func (cmd *GetUsers) Event() Event {
	e := eventGetUsers(cmd.users)
	return &e
}

type eventGetUsers []*core.User

func (evt *eventGetUsers) Event() string {
	return successName(CommandGetUsers)
}

func (evt *eventGetUsers) Code() int {
	return http.StatusOK
}
