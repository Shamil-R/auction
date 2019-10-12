package command

import (
	"gitlab/nefco/auction/core"
	"net/http"
)

const CommandGetUser = "get.user"

type GetUser struct {
	*userCommand
	user *core.User
}

func newGetUser(user *core.User) *GetUser {
	return &GetUser{
		userCommand: newUserCommand(CommandGetUser, AccessAll, user),
	}
}

func (cmd *GetUser) SetUser(user *core.User) {
	cmd.user = user
}

func (cmd *GetUser) Event() Event {
	return &struct {
		*event
		*core.User
	}{
		newSucces(cmd.name, http.StatusOK),
		cmd.user,
	}
}
