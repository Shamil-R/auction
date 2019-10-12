package command

import (
	"gitlab/nefco/auction/core"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
)

const CommandDeleteUserGroup = "delete.user_group"

type DeleteUserGroup struct {
	*base
	UserID   uint   `validate:"required"`
	GroupKey string `validate:"required"`
}

func newDeleteUserGroup(user *core.User) *DeleteUserGroup {
	return &DeleteUserGroup{
		base: newBase(CommandDeleteUserGroup, AccessRoot, user),
	}
}

func (cmd *DeleteUserGroup) GetUserID() uint {
	return cmd.UserID
}

func (cmd *DeleteUserGroup) GetGroupKey() string {
	return cmd.GroupKey
}

func (cmd *DeleteUserGroup) Event() Event {
	return newSucces(cmd.name, http.StatusNoContent)
}

func (cmd *DeleteUserGroup) eject(ctx echo.Context) error {
	userID, err := strconv.ParseUint(ctx.Param("userID"), 10, 32)
	if err != nil {
		return err
	}
	cmd.UserID = uint(userID)
	cmd.GroupKey = ctx.Param("groupKey")
	return nil
}
