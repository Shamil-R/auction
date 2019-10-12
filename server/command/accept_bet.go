package command

import (
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/errors"
	"strconv"

	"github.com/labstack/echo"
)

const CommandAcceptBet = "accept.bet"

var missingPathParamBetID = errors.BadRequest("Missing path param 'betID'")

type AcceptBet struct {
	*lot
	BetId uint64 `json:"bet_id" validate:"required"`
	l     *core.Lot
}

func newAcceptBet(user *core.User) *AcceptBet {
	return &AcceptBet{
		lot: newLot(CommandAcceptBet, AccessManager, user),
	}
}

func (cmd *AcceptBet) BetID() uint64 {
	return cmd.BetId
}

func (cmd *AcceptBet) GetLot() *core.Lot {
	return cmd.l
}

func (cmd *AcceptBet) SetLot(lot *core.Lot) {
	cmd.l = lot
}

func (cmd *AcceptBet) eject(ctx echo.Context) error {
	betID, err := strconv.Atoi(ctx.Param("betID"))
	if err != nil {
		return missingPathParamBetID
	}
	cmd.BetId = uint64(betID)
	return nil
}
