package command

import (
	"gitlab/nefco/auction/core"
	"net/http"

	"github.com/labstack/echo"
)

const CommandAddGroup = "add.group"

type AddGroup struct {
	*base
	*core.Group
}

func newAddGroup(user *core.User) *AddGroup {
	return &AddGroup{
		base:  newBase(CommandAddGroup, AccessRoot, user),
		Group: &core.Group{},
	}
}

func (cmd *AddGroup) GetGroup() *core.Group {
	return cmd.Group
}

func (cmd *AddGroup) Event() Event {
	return &struct {
		*event
		*core.Group
	}{
		newSucces(cmd.name, http.StatusCreated),
		cmd.Group,
	}
}

func (cmd *AddGroup) eject(ctx echo.Context) error {
	return ctx.Bind(cmd)
}
