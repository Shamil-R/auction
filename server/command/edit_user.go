package command

import (
	"gitlab/nefco/auction/core"
	"net/http"

	"github.com/labstack/echo"
)

const CommandEditUser = "edit.user"

type EditUser struct {
	*base
	*core.UserInfo
}

func newEditUser(user *core.User) *EditUser {
	return &EditUser{
		base:     newBase(CommandEditUser, AccessUser, user),
		UserInfo: &core.UserInfo{},
	}
}

func (cmd *EditUser) GetUser() *core.UserInfo {
	return cmd.UserInfo
}

func (cmd *EditUser) Event() Event {
	return &struct {
		*event
		*core.UserInfo
	}{
		newSucces(cmd.name, http.StatusCreated),
		cmd.UserInfo,
	}
}

func (cmd *EditUser) eject(ctx echo.Context) error {
	return ctx.Bind(cmd)
}
