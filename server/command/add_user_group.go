package command

import (
	"gitlab/nefco/auction/core"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
)

const CommandAddUserGroup = "add.user_group"

type AddUserGroup struct {
	*base
	UserID   uint   `validate:"required"`
	GroupKey string `validate:"required"`
}

func newAddUserGroup(user *core.User) *AddUserGroup {
	return &AddUserGroup{
		base: newBase(CommandAddUserGroup, AccessRoot, user),
	}
}

func (cmd *AddUserGroup) GetUserID() uint {
	return cmd.UserID
}

func (cmd *AddUserGroup) GetGroupKey() string {
	return cmd.GroupKey
}

func (cmd *AddUserGroup) Event() Event {
	return newSucces(cmd.name, http.StatusNoContent)
}

func (cmd *AddUserGroup) eject(ctx echo.Context) error {
	userID, err := strconv.ParseUint(ctx.Param("userID"), 10, 32)
	if err != nil {
		return err
	}
	cmd.UserID = uint(userID)
	cmd.GroupKey = ctx.Param("groupKey")
	return nil
}
