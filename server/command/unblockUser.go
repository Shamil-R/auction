package command

import (
	"github.com/labstack/echo"
	"gitlab/nefco/auction/core"
	"net/http"
	"strconv"
)

const CommandUnblockUser = "unblock.user"

type UnblockUser struct {
	*base
	UserID uint `validate:"required"`
}

func NewUnblockUser(user *core.User) *UnblockUser {
	return &UnblockUser{
		base: newBase(CommandUnblockUser, AccessManager, user),
	}
}

func (cmd *UnblockUser) eject(ctx echo.Context) error {
	userID, err := strconv.ParseUint(ctx.Param("userID"), 10, 32)
	if err != nil {
		return err
	}
	cmd.UserID = uint(userID)
	return nil
}


func (cmd *UnblockUser) Event() Event {
	return newSucces(cmd.name, http.StatusNoContent)
}
