package command

import (
	"github.com/labstack/echo"
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/errors"
	"strconv"
)

const CommandAllowChangeAct = "allow_change_act"

type AllowActEditing struct {
	*lot
	AllowChange int
}

func (cmd *AllowActEditing) GetAllowChange() int {
	return cmd.AllowChange
}

func newAllowChangeAct(user *core.User) *AllowActEditing {
	return &AllowActEditing{
		lot: newLot(CommandAllowChangeAct, AccessManager, user),
	}
}

func (cmd *AllowActEditing) eject(ctx echo.Context) error {
	allowChange := &struct {
		Value string `json:"allow_change" validate:"required"`
	}{}

	err :=  ctx.Bind(allowChange)
	if err != nil {
		return err
	}

	i, err := strconv.Atoi(allowChange.Value)
	if err != nil {
		return err
	}

	if i > 1 {
		return errors.BadRequest("Хули тут так много?* ")
	}

	cmd.AllowChange = i
	return nil
}