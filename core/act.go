package core

import (
	lib "gitlab/nefco/auction/lib"
)

type Act struct {
	ObjectID    uint         `json:"-"`
	ActNumber   string       `json:"act_number" validate:"required"`
	Date        lib.DateTime `json:"act_date" validate:"required"`
	AllowChange bool         `json:"allow_change"`
}

func (a *Act) ObjectId() uint {
	return a.ObjectID
}

func (a *Act) DocDate() lib.DateTime {
	return a.Date
}

func (a *Act) DocNumber() string {
	return a.ActNumber
}
