package command

import (
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/core/object"

	"github.com/labstack/echo"
)

const CommandCompleteLot = "complete.lot"

type CompleteLot struct {
	*lot
	ConfirmInfo object.JSONData `json:"info"`
}

func newCompleteLot(user *core.User) *CompleteLot {
	return &CompleteLot{
		lot: newLot(CommandCompleteLot, AccessManagerUser, user),
	}
}

func (cmd *CompleteLot) Info() object.JSONData {
	return cmd.ConfirmInfo
}

func (cmd *CompleteLot) eject(ctx echo.Context) error {
	return ctx.Bind(cmd)
}
