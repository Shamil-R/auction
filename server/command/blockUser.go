package command

import (
	"github.com/labstack/echo"
	"gitlab/nefco/auction/core"
	"net/http"
	"strconv"
)

const CommandBlockUser = "block.user"

type BlockUser struct {
	*base
	UserID uint `validate:"required"`
}

func NewBlockUser(user *core.User) *BlockUser {
	return &BlockUser{
		base: newBase(CommandBlockUser, AccessManager, user),
	}
}

func (cmd *BlockUser) eject(ctx echo.Context) error {
	userID, err := strconv.ParseUint(ctx.Param("userID"), 10, 32)
	if err != nil {
		return err
	}
	cmd.UserID = uint(userID)
	return nil
}

func (cmd *BlockUser) Event() Event {
	return newSucces(cmd.name, http.StatusNoContent)
}
