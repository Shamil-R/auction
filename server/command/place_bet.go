package command

import (
	"gitlab/nefco/auction/core"

	"github.com/labstack/echo"
)

const CommandPlaceBet = "place.bet"

type PlaceBet struct {
	*lot
	l        *core.Lot
	BetValue uint `json:"value"`
}

func newPlaceBet(user *core.User) *PlaceBet {
	return &PlaceBet{
		lot: newLot(CommandPlaceBet, AccessUser, user),
	}
}

func (cmd *PlaceBet) Value() uint {
	return cmd.BetValue
}

func (cmd *PlaceBet) GetLot() *core.Lot {
	return cmd.l
}

func (cmd *PlaceBet) SetLot(lot *core.Lot) {
	cmd.l = lot
}

func (cmd *PlaceBet) eject(ctx echo.Context) error {
	return ctx.Bind(cmd)
}
