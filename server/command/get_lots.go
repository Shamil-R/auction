package command

import (
	"encoding/json"
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/core/object"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
)

const CommandGetLots = "get.lots"

type GetLotsFilter struct {
	executor           *core.User
	FilterGroup        string `json:"group_key" validate:"required"`
	FilterState        string `json:"state" validate:"state"`
	FilterLotID        uint   `json:"lot_id"`
	FilterObjectID     uint   `json:"object_id"`
	FilterStartPrice   uint   `json:"start_price"`
	FilterEndPrice     uint   `json:"end_price"`
	FilterOnlyWithBets bool   `json:"only_with_bets"`
}

type GetLots struct {
	*base
	GetLotsFilter
	ObjectFilter object.ObjectFilter
	lots         []*core.Lot
}

func newGetLots(user *core.User) *GetLots {
	//TODO: у root возникает ошибка при вызове ObjectFilter(), т.к. user.ObjectType = nil
	cmd := &GetLots{
		base: newBase(CommandGetLots, AccessManagerUser, user),
		GetLotsFilter: GetLotsFilter{
			executor: user,
		},
		ObjectFilter: user.ObjectType.ObjectFilter(),
	}

	user.SetLotsFilter(cmd)

	return cmd
}

func (cmd *GetLots) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &cmd.GetLotsFilter); err != nil {
		return err
	}
	if err := json.Unmarshal(b, cmd.ObjectFilter); err != nil {
		return err
	}
	return nil
}

func (cmd *GetLots) Groups() []string {
	groups := make([]string, len(cmd.user.Groups))
	for i, group := range cmd.user.Groups {
		groups[i] = group.Key
	}
	return groups
}

func (cmd *GetLots) Executor() *core.User {
	return cmd.executor
}

func (cmd *GetLots) Group() string {
	return cmd.FilterGroup
}

func (cmd *GetLots) State() string {
	return cmd.FilterState
}

func (cmd *GetLots) LotID() uint {
	return cmd.FilterLotID
}

func (cmd *GetLots) ObjectID() uint {
	return cmd.FilterObjectID
}

func (cmd *GetLots) StartPrice() uint {
	return cmd.FilterStartPrice
}

func (cmd *GetLots) EndPrice() uint {
	return cmd.FilterEndPrice
}

func (cmd *GetLots) OnlyWithBets() bool {
	return cmd.FilterOnlyWithBets
}

func (cmd *GetLots) Filter() object.ObjectFilter {
	return cmd.ObjectFilter
}

func (cmd *GetLots) SetLots(lots []*core.Lot) {
	cmd.lots = lots
}

func (cmd *GetLots) Event() Event {
	e := eventGetLots(cmd.lots)
	return &e
}

func (cmd *GetLots) eject(ctx echo.Context) error {
	cmd.FilterGroup = ctx.QueryParam("group_key")
	cmd.FilterState = ctx.QueryParam("state")
	paramObjectID := ctx.QueryParam("object_id")
	if len(paramObjectID) > 0 {
		objectID, err := strconv.ParseUint(paramObjectID, 10, 32)
		if err != nil {
			return err
		}
		cmd.FilterObjectID = uint(objectID)
	}
	paramStartPrice := ctx.QueryParam("start_price")
	if len(paramObjectID) > 0 {
		startPrice, err := strconv.ParseUint(paramStartPrice, 10, 32)
		if err != nil {
			return err
		}
		cmd.FilterStartPrice = uint(startPrice)
	}
	paramEndPrice := ctx.QueryParam("end_price")
	if len(paramObjectID) > 0 {
		endPrice, err := strconv.ParseUint(paramEndPrice, 10, 32)
		if err != nil {
			return err
		}
		cmd.FilterEndPrice = uint(endPrice)
	}
	if err := cmd.ObjectFilter.Fill(ctx.QueryParams()); err != nil {
		return err
	}
	cmd.FilterOnlyWithBets = ctx.QueryParam("only_with_bets") == "true"
	return nil
}

type eventGetLots []*core.Lot

func (evt *eventGetLots) Event() string {
	return successName(CommandGetLots)
}

func (evt *eventGetLots) Code() int {
	return http.StatusOK
}
