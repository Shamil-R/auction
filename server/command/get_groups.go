package command

import (
	"gitlab/nefco/auction/core"
	"net/http"
)

const CommandGetGroups = "get.groups"

type GetGroups struct {
	*base
	groups []*core.Group
}

func newGetGroups(user *core.User) *GetGroups {
	return &GetGroups{
		base: newBase(CommandGetGroups, AccessManagerUser, user),
	}
}

func (cmd *GetGroups) ObjectType() string {
	return cmd.user.ObjectType.Type()
}

func (cmd *GetGroups) SetGroups(groups []*core.Group) {
	cmd.groups = groups
}

func (cmd *GetGroups) Event() Event {
	e := eventGetGroups(cmd.groups)
	return &e
}

type eventGetGroups []*core.Group

func (evt *eventGetGroups) Event() string {
	return successName(CommandGetGroups)
}

func (evt *eventGetGroups) Code() int {
	return http.StatusOK
}
