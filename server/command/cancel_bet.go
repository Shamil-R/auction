package command

import "gitlab/nefco/auction/core"

const CommandCancelBet = "cancel.bet"

type CancelBet struct {
	*lot
	l *core.Lot
}

func newCancelBet(user *core.User) *CancelBet {
	return &CancelBet{
		lot: newLot(CommandCancelBet, AccessUser, user),
	}
}

func (cmd *CancelBet) GetLot() *core.Lot {
	return cmd.l
}

func (cmd *CancelBet) SetLot(lot *core.Lot) {
	cmd.l = lot
}
