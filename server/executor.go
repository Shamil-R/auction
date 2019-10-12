package server

import (
	"gitlab/nefco/auction/auction"
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/core/object"
	"gitlab/nefco/auction/errors"
	"gitlab/nefco/auction/process/rule"
	"gitlab/nefco/auction/server/command"
	"net/http"

	"github.com/liip/sheriff"
	"go.uber.org/zap"
	validator "gopkg.in/go-playground/validator.v9"
)

var (
	commandNotSupported = errors.BadRequest("Command not supported")
	forbidden           = errors.Forbidden("Forbidden")
	validationFailed    = errors.NewError("Validation failed", http.StatusUnprocessableEntity)
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	core.ValidateState(validate)
	object.ValidateObjectType(validate)
	rule.ValidateRule(validate)
}

func execute(cmd command.Command, auction auction.Auction) (command.Event, error) {
	logger := zap.L().Named("execute").With(
		zap.String("command", cmd.Command()),
		zap.Uint("user_id", cmd.Executor().ID),
	)

	if !cmd.Access() {
		logger.Warn("forbidden")
		return nil, forbidden
	}

	if err := validate.Struct(cmd); err != nil {
		logger.Warn("command validation failed", zap.Error(err))
		return nil, validationFailed
	}

	var err error

	switch c := cmd.(type) {
	case *command.GetGroups:
		err = auction.Groups(c)
	case *command.AddGroup:
		err = auction.CreateGroup(c)
	case *command.GetUsers:
		err = auction.Users(c)
	case *command.AddUser:
		err = auction.CreateUser(c)
	case *command.EditUser:
		err = auction.EditUser(c)
	case *command.GetUser:
		err = auction.GetUser(c)
	case *command.AddUserGroup:
		err = auction.AddUserGroup(c)
	case *command.DeleteUserGroup:
		err = auction.DeleteUserGroup(c)
	case *command.BlockUser:
		err = auction.BlockUser(c.UserID)
	case *command.UnblockUser:
		err = auction.UnblockUser(c.UserID)
	case *command.GetLots:
		err = auction.Lots(c)
	case *command.AddLot:
		err = auction.CreateLot(c)
	case *command.AutoBooking:
		err = auction.AutoBookingLot(c)
	case *command.GetLot:
		err = auction.GetLot(c)
	case *command.EditLot:
		err = auction.UpdateLot(c)
	case *command.DeleteLot:
		err = auction.DeleteLot(c)
	case *command.PlaceBet:
		err = auction.PlaceBet(c)
	case *command.CancelBet:
		err = auction.CancelBet(c)
	case *command.ConfirmLot:
		err = auction.ConfirmLot(c)
	case *command.DeleteConfirmation:
		err = auction.DeleteConfirmation(c)
	case *command.CompleteLot:
		err = auction.CompleteLot(c)
	case *command.GetHistory:
		err = auction.History(c)
	case *command.AcceptBet:
		err = auction.AcceptBet(c)
	case *command.SendFeedback:
		err = auction.SendFeedback(c)
	case *command.GetBookingWindow:
		err = auction.GetBookingWindow(c)
	case *command.EditAct:
		err = auction.UpdateLotAct(c)
	case *command.AllowActEditing:
		err = auction.AllowChangeAct(c)

	default:
		logger.Warn("command not supported")
		return nil, commandNotSupported
	}

	if err != nil {
		logger.Error("execution failed", zap.Error(err))
		return nil, err
	}

	return cmd.Event(), nil
}

func response(executor *core.User, evt command.Event) (interface{}, error) {
	options := &sheriff.Options{
		Groups:      []string{executor.Level()},
		ShowDefault: true,
	}

	return sheriff.Marshal(options, evt)
}
