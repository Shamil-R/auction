package service

import (
	"database/sql"
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/db"
)

type groupService struct {
	tx *db.Tx
}

func NewGroupService(tx *db.Tx) *groupService {
	return &groupService{tx}
}

func (svc *groupService) Groups(filter core.ObjectTypeFilter) ([]*core.Group, error) {
	groups := []*core.Group{}

	arg := map[string]interface{}{
		"object_type": filter.ObjectType(),
	}

	query := `
		SELECT * FROM groups WHERE object_type = :object_type
	`

	if err := svc.tx.Select(&groups, query, arg); err != nil {
		return nil, err
	}

	return groups, nil
}

func (svc *groupService) GroupByKey(groupKey string) (*core.Group, error) {
	group := &core.Group{}

	arg := map[string]interface{}{
		"group_key": groupKey,
	}

	query := `
		SELECT * FROM groups WHERE group_key = :group_key
	`

	if err := svc.tx.Get(group, query, arg); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return group, nil
}

func (svc *groupService) CreateGroup(group *core.Group) error {
	query := `
		INSERT INTO groups (
			group_key, 
			name,
			object_type
		)
		VALUES (
			:group_key, 
			:name,
			:object_type
		)
	`

	_, err := svc.tx.Exec(query, group)
	if err != nil {
		return err
	}

	return nil
}
