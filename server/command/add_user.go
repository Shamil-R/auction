package command

import (
	"gitlab/nefco/auction/core"
	"net/http"

	"github.com/labstack/echo"
)

const CommandAddUser = "add.user"

type AddUser struct {
	*base
	*core.User
}

func newAddUser(user *core.User) *AddUser {
	return &AddUser{
		base: newBase(CommandAddUser, AccessRoot, user),
		User: &core.User{},
	}
}

func (cmd *AddUser) GetUser() *core.User {
	return cmd.User
}

func (cmd *AddUser) Event() Event {
	return &struct {
		*event
		*core.User
	}{
		newSucces(cmd.name, http.StatusCreated),
		cmd.User,
	}
}

func (cmd *AddUser) eject(ctx echo.Context) error {
	return ctx.Bind(cmd)
}
