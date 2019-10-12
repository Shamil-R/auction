package command

import (
	"github.com/labstack/echo"
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/errors"
)

const CommandEditAct = "edit.act"

type EditAct struct {
	*lot
	*core.Act
}

func newEditAct(user *core.User) *EditAct {
	return &EditAct{
		lot: newLot(CommandEditAct, AccessUser, user),
		Act: &core.Act{},
	}
}

func(cmd *EditAct) eject(ctx echo.Context) error {
	if err := ctx.Bind(cmd); err != nil {
		return err
	}
	if cmd.Act.ActNumber == "" {
		return errors.BadRequest("the doc_number cannot be empty!")
	}
	if cmd.Act.Date == "" {
		return errors.BadRequest("the date cannot be empty!")
	}
	return cmd.Act.Date.Validate()
}

func (cmd *EditAct) GetObjectAct() *core.Act {
	return cmd.Act
}
