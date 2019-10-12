package core

import "gitlab/nefco/auction/core/object"

type Group struct {
	Key        string            `json:"key" db:"group_key" validate:"required"`
	Name       string            `json:"name" db:"name" validate:"required"`
	ObjectType object.ObjectType `json:"object_type" db:"object_type" validate:"required,object_type"`
}

type ObjectTypeFilter interface {
	ObjectType() string
}

type GroupService interface {
	Groups(filter ObjectTypeFilter) ([]*Group, error)
	GroupByKey(groupKey string) (*Group, error)
	CreateGroup(group *Group) error
}
