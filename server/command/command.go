package command

import (
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/errors"
	"net/http"

	"strconv"
	"strings"

	"github.com/labstack/echo"
)

var (
	AccessRoot        = []string{core.LevelRoot}
	AccessManager     = []string{core.LevelManager}
	AccessUser        = []string{core.LevelUser}
	AccessManagerUser = []string{core.LevelManager, core.LevelUser}
	AccessAll         = []string{core.LevelRoot, core.LevelManager, core.LevelUser}
)

var (
	missingPathParamUserID = errors.BadRequest("Missing path param 'userID'")
	missingPathParamLotID  = errors.BadRequest("Missing path param 'lotID'")
)

type eject interface {
	eject(ctx echo.Context) error
}

type ejectLot interface {
	ejectLot(ctx echo.Context) error
}

type Command interface {
	Command() string
	Access() bool
	Executor() *core.User
	Event() Event
	commandName() string
}

type base struct {
	name   string
	levels []string
	user   *core.User
}

func newBase(name string, levels []string, user *core.User) *base {
	return &base{name, levels, user}
}

func (cmd *base) Command() string {
	return commandName(cmd.name)
}

func (cmd *base) Access() bool {
	for _, l := range cmd.levels {
		if cmd.user.Level() == l {
			return true
		}
	}
	return false
}

func (cmd *base) Executor() *core.User {
	return cmd.user
}

func (cmd *base) Event() Event {
	return newSucces(cmd.name, http.StatusNoContent)
}

func (cmd *base) commandName() string {
	return cmd.name
}

type userCommand struct {
	*base
	UserId uint `json:"user_id" validate:"required"`
}

func newUserCommand(name string, levels []string, user *core.User) *userCommand {
	return &userCommand{
		base: newBase(name, levels, user),
	}
}

func (cmd *userCommand) UserID() uint {
	return cmd.UserId
}

func (cmd *userCommand) ejectLot(ctx echo.Context) error {
	userID, err := strconv.Atoi(ctx.Param("userID"))
	if err != nil {
		return missingPathParamUserID
	}
	cmd.UserId = uint(userID)
	return nil
}

type lot struct {
	*base
	LotId uint `json:"lot_id" validate:"required"`
}

func newLot(name string, levels []string, user *core.User) *lot {
	return &lot{
		base: newBase(name, levels, user),
	}
}

func (cmd *lot) LotID() uint {
	return cmd.LotId
}

func (cmd *lot) ejectLot(ctx echo.Context) error {
	lotID, err := strconv.Atoi(ctx.Param("lotID"))
	if err != nil {
		return missingPathParamLotID
	}
	cmd.LotId = uint(lotID)
	return nil
}

type Event interface {
	Event() string
	Code() int
}

type event struct {
	name string
	code int
}

func (evt *event) Event() string {
	return evt.name
}

func (evt *event) Code() int {
	return evt.code
}

func newEvent(name string, code int) *event {
	return &event{name, code}
}

func newSucces(name string, code int) *event {
	return newEvent(successName(name), code)
}

func LotEvent(name string, lot *core.Lot) Event {
	return &struct {
		*event
		*core.Lot
	}{newEvent(name, http.StatusOK), lot}
}

func ErrorEvent(name string, err error) Event {
	return &struct {
		*event
		error
	}{
		event: newEvent(name, 0),
		error: err,
	}
}

func Fail(cmd Command, err error) Event {
	return &struct {
		*event
		error
	}{
		event: &event{failedName(cmd.commandName()), 0},
		error: err,
	}
}

func New(name string, user *core.User) Command {
	switch name {
	case CommandGetGroups:
		return newGetGroups(user)
	case CommandAddGroup:
		return newAddGroup(user)
	case CommandGetUsers:
		return newGetUsers(user)
	case CommandAddUser:
		return newAddUser(user)
	case CommandGetUser:
		return newGetUser(user)
	case CommandEditUser:
		return newEditUser(user)
	case CommandAddUserGroup:
		return newAddUserGroup(user)
	case CommandDeleteUserGroup:
		return newDeleteUserGroup(user)
	case CommandBlockUser:
		return NewBlockUser(user)
	case CommandUnblockUser:
		return NewUnblockUser(user)
	case CommandGetLots:
		return newGetLots(user)
	case CommandAddLot:
		return newAddLot(user)
	case CommandGetLot:
		return newGetLot(user)
	case CommandEditLot:
		return newEditLot(user)
	case CommandDeleteLot:
		return newDeleteLot(user)
	case CommandPlaceBet:
		return newPlaceBet(user)
	case CommandCancelBet:
		return newCancelBet(user)
	case CommandConfirmLot:
		return newConfirmLot(user)
	case CommandDeleteConfirmation:
		return newDeleteConfirmation(user)
	case CommandCompleteLot:
		return newCompleteLot(user)
	case CommandPlaceReservation:
		return newPlaceReservation(user)
	case CommandCancelReservation:
		return newCancelReservation(user)
	case CommandGetHistory:
		return newGetHistory(user)
	case CommandAcceptBet:
		return newAcceptBet(user)
	case CommandSendFeedback:
		return newSendFeedback(user)
	case CommandGetBookingWindow:
		return newGetBookingWindow(user)
	// act commanders:
	case CommandEditAct:
		return newEditAct(user)
	case CommandAutoBooking:
		return newAutoBooking(user)
	case CommandAllowChangeAct:
		return newAllowChangeAct(user)
	}
	return nil
}

func Fill(cmd Command, ctx echo.Context) error {
	if eject, ok := cmd.(ejectLot); ok {
		if err := eject.ejectLot(ctx); err != nil {
			return err
		}
	}
	if eject, ok := cmd.(eject); ok {
		if err := eject.eject(ctx); err != nil {
			return err
		}
	}
	return nil
}

func ExtractCommand(commandType string) string {
	return strings.TrimPrefix(commandType, "command.")
}

func commandName(name string) string {
	return "command." + name
}

func eventName(name string) string {
	return "event." + name
}

func successName(name string) string {
	return eventName(name) + ".success"
}

func failedName(name string) string {
	return eventName(name) + ".failed"
}
