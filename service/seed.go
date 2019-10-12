package service

import (
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/core/object"
	"gitlab/nefco/auction/db"
)

type userGroup struct {
	userID   uint
	groupKey string
}

func (ug *userGroup) GetUserID() uint {
	return ug.userID
}
func (ug *userGroup) GetGroupKey() string {
	return ug.groupKey
}

func Seed(db *db.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if err := seed(tx); err != nil {
		if err := tx.Rollback(); err != nil {
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func seed(tx *db.Tx) error {
	// if err := groupSeed(NewGroupService(tx)); err != nil {
	// 	return err
	// }
	// if err := userSeed(NewUserService(tx)); err != nil {
	// 	return err
	// }
	if err := moveUsers(tx); err != nil {
		return err
	}
	return nil
}

func groupSeed(svc core.GroupService) error {
	group, err := svc.GroupByKey("nc_trip")
	if err != nil {
		return err
	}
	if group == nil {
		group := &core.Group{
			Key:        "nc_trip",
			Name:       "НЭФИС Косметикс",
			ObjectType: object.NewObjectType(object.ObjectTypeTrip),
		}
		if err := svc.CreateGroup(group); err != nil {
			return err
		}
	}
	group, err = svc.GroupByKey("nbp_trip")
	if err != nil {
		return err
	}
	if group == nil {
		group := &core.Group{
			Key:        "nbp_trip",
			Name:       "НЭФИС-БИОПРОДУКТ",
			ObjectType: object.NewObjectType(object.ObjectTypeTrip),
		}
		if err := svc.CreateGroup(group); err != nil {
			return err
		}
	}
	return nil
}

func userSeed(svc core.UserService) error {
	_, err := svc.UserByUsername("manager")
	if err != nil && err != core.UserNotFound {
		return err
	}

	if err == nil {
		return nil
	}

	password, _ := core.HashPassword("password")
	// manager
	manager := &core.User{
		Username:   "manager",
		Password:   password,
		ObjectType: object.NewObjectType(object.ObjectTypeTrip),
		BackKey:    "secret_key",
	}
	if err := svc.CreateUser(manager); err != nil {
		return err
	}
	if err := svc.AddGroup(&userGroup{manager.ID, "nc_trip"}); err != nil {
		return nil
	}
	if err := svc.AddGroup(&userGroup{manager.ID, "nbp_trip"}); err != nil {
		return nil
	}
	// user3
	user1 := &core.User{
		Username:   "user1",
		Password:   password,
		ObjectType: object.NewObjectType(object.ObjectTypeTrip),
	}
	if err := svc.CreateUser(user1); err != nil {
		return err
	}
	if err := svc.AddGroup(&userGroup{user1.ID, "nc_trip"}); err != nil {
		return nil
	}
	if err := svc.AddGroup(&userGroup{user1.ID, "nbp_trip"}); err != nil {
		return nil
	}
	// user2
	user2 := &core.User{
		Username:   "user2",
		Password:   password,
		ObjectType: object.NewObjectType(object.ObjectTypeTrip),
	}
	if err := svc.CreateUser(user2); err != nil {
		return err
	}
	if err := svc.AddGroup(&userGroup{user2.ID, "nc_trip"}); err != nil {
		return nil
	}
	if err := svc.AddGroup(&userGroup{user2.ID, "nbp_trip"}); err != nil {
		return nil
	}
	// user3
	user3 := &core.User{
		Username:   "user3",
		Password:   password,
		ObjectType: object.NewObjectType(object.ObjectTypeTrip),
	}
	if err := svc.CreateUser(user3); err != nil {
		return err
	}
	if err := svc.AddGroup(&userGroup{user3.ID, "nc_trip"}); err != nil {
		return nil
	}
	// user4
	user4 := &core.User{
		Username:   "user4",
		Password:   password,
		ObjectType: object.NewObjectType(object.ObjectTypeTrip),
	}
	if err := svc.CreateUser(user4); err != nil {
		return err
	}
	if err := svc.AddGroup(&userGroup{user4.ID, "nbp_trip"}); err != nil {
		return nil
	}
	return nil
}

func moveUsers(tx *db.Tx) error {
	contractors := []struct {
		Login      string `db:"login"`
		Password   string `db:"password"`
		Factory    int    `db:"factory"`
		Contractor int    `db:"contractor"`
	}{}

	query := `
		SELECT 
			ct.login AS login, 
			MIN(ct.password) AS password,
			SUM(c.factory_id) AS factory,
			MIN(ct.id) AS contractor
		FROM nefco.dbo.co_contractor_attr_transp AS ct
		JOIN nefco.dbo.co_contractor AS c 
			ON c.id = ct.contractor_id
		WHERE 
			(ct.login IS NOT NULL AND ct.password IS NOT NULL)
		AND c.active = 1
		GROUP BY login
		ORDER BY factory DESC
	`

	if err := tx.Select(&contractors, query, nil); err != nil {
		return err
	}

	svc := NewUserService(tx)

	for _, c := range contractors {
		_, err := svc.UserByUsername(c.Login)
		if err != nil && err != core.UserNotFound {
			return err
		}

		if err == nil {
			continue
		}

		password, err := core.HashPassword(c.Password)
		if err != nil {
			return err
		}

		user := &core.User{
			Username:   c.Login,
			Password:   password,
			ObjectType: object.NewObjectType(object.ObjectTypeTrip),
		}

		if err := svc.CreateUser(user); err != nil {
			return err
		}

		if c.Factory == 1 || c.Factory == 3 {
			arg := map[string]interface{}{
				"user_id":   user.ID,
				"group_key": "nc_trip",
			}

			query := `
				INSERT INTO users_groups (user_id, group_key)
				VALUES (:user_id, :group_key)
			`

			if _, err := svc.tx.Exec(query, arg); err != nil {
				return err
			}
		}

		if c.Factory == 2 || c.Factory == 3 {
			arg := map[string]interface{}{
				"user_id":   user.ID,
				"group_key": "nbp_trip",
			}

			query := `
				INSERT INTO users_groups (user_id, group_key)
				VALUES (:user_id, :group_key)
			`

			if _, err := svc.tx.Exec(query, arg); err != nil {
				return err
			}
		}

		arg := map[string]interface{}{
			"user_id":       user.ID,
			"contractor_id": c.Contractor,
		}

		query := `
			UPDATE nefco.dbo.co_contractor_attr_transp
			SET auction_user_id = :user_id
			WHERE id = :contractor_id
		`

		if _, err := svc.tx.Exec(query, arg); err != nil {
			return err
		}
	}

	return nil
}
