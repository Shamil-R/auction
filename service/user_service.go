package service

import (
	"database/sql"
	"fmt"
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/db"
	"gitlab/nefco/auction/errors"
)

type userService struct {
	tx *db.Tx
}

func NewUserService(tx *db.Tx) *userService {
	return &userService{tx}
}

func (svc *userService) Users() ([]*core.User, error) {
	users := []*core.User{}

	query := `
		SELECT id, username, object_type, back_key
		FROM users 
	`

	if err := svc.tx.Select(&users, query, nil); err != nil {
		return nil, err
	}

	return users, nil
}

func (svc *userService) Managers() ([]*core.User, error) {
	users := []*core.User{}

	query := `
		SELECT id, username, object_type, back_key
		FROM users 
		WHERE back_key IS NOT NULL
		AND back_key != ''
	`

	if err := svc.tx.Select(&users, query, nil); err != nil {
		return nil, err
	}

	return users, nil
}

func (svc *userService) User(id uint) (*core.User, error) {
	user := &core.User{}

	arg := map[string]interface{}{
		"id": id,
	}

	query := `
		SELECT
			id,
			username,
			blocked,
			object_type,
			back_key,
			employer_name,
			employer_surname,
			employer_patronymic,
			phone_number,
			company_name
		FROM users 
		WHERE id = :id
	`

	if err := svc.tx.Get(user, query, arg); err != nil {
		if err == sql.ErrNoRows {
			return nil, core.UserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (svc *userService) UserByUsername(username string) (*core.User, error) {
	user := &core.User{}

	arg := map[string]interface{}{
		"username": username,
	}

	query := `
		SELECT * FROM users 
		WHERE username = :username
	`

	if err := svc.tx.Get(user, query, arg); err != nil {
		if err == sql.ErrNoRows {
			return nil, core.UserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (svc *userService) CreateUser(user *core.User) error {
	query := `
		INSERT INTO users (
			username, 
			password,
			object_type,
			back_key,
			employer_name,
			employer_surname,
			employer_patronymic,
			company_name,
			phone_number
		)
		VALUES (
			:username, 
			:password,
			:object_type,
			:back_key,
			:employer_name,
			:employer_surname,
			:employer_patronymic,
			:company_name,
			:phone_number
		)
	`

	result, err := svc.tx.Exec(query, user)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = uint(id)

	return nil
}

func (svc *userService) SaveUser(user *core.User) error {
	query := `
		UPDATE users SET
			employer_name = :employer_name,
			employer_surname = :employer_surname,
			employer_patronymic = :employer_patronymic,
			company_name = :company_name,
			phone_number = :phone_number
		WHERE id = :id
	`

	if _, err := svc.tx.Exec(query, user); err != nil {
		return err
	}

	return nil
}

func (svc *userService) UpdateUserPassword(userID uint, password string) error {
	query := `UPDATE users SET password = :password WHERE id = :id`

	arg := map[string]interface{}{
		"id":       userID,
		"password": password,
	}

	if _, err := svc.tx.Exec(query, arg); err != nil {
		return err
	}

	return nil
}

func (svc *userService) AddGroup(ug core.UserGroup) error {
	arg := map[string]interface{}{
		"user_id":   ug.GetUserID(),
		"group_key": ug.GetGroupKey(),
	}

	query := `
		INSERT INTO users_groups (user_id, group_key)
		VALUES (:user_id, :group_key)
	`

	if _, err := svc.tx.Exec(query, arg); err != nil {
		return err
	}

	return nil
}

func (svc *userService) DeleteGroup(ug core.UserGroup) error {
	arg := map[string]interface{}{
		"user_id":   ug.GetUserID(),
		"group_key": ug.GetGroupKey(),
	}

	query := `
		DELETE FROM users_groups 
		WHERE user_id = :user_id 
		AND group_key = :group_key
	`

	if _, err := svc.tx.Exec(query, arg); err != nil {
		return err
	}

	return nil
}

func (svc *userService) GroupsByUser(userID uint) ([]*core.Group, error) {
	groups := []*core.Group{}

	arg := map[string]interface{}{
		"user_id": userID,
	}

	query := `
		SELECT g.* 
		FROM users_groups AS ug
		JOIN groups AS g ON g.group_key = ug.group_key
		WHERE ug.user_id = :user_id
	`

	if err := svc.tx.Select(&groups, query, arg); err != nil {
		return nil, err
	}

	return groups, nil
}

func (svc *userService) BlockUser(userID uint) error {

	if err := validateBlockParams(userID, svc.tx); err != nil {
		return err
	}

	return changeUserBlocked(userID, true, svc.tx)

}

func (svc *userService) UnblockUser(userID uint) error {

	if err := validateBlockParams(userID, svc.tx); err != nil {
		return err
	}

	return changeUserBlocked(userID, false, svc.tx)
}

func validateBlockParams(userID uint, tx *db.Tx) error {
	errMessage := "can't block/unblock this user"
	// if user is admin:
	if userID == core.RootUserID {
		return errors.BadRequest(errMessage)
	}

	arg := map[string]interface{}{
		"id": userID,
	}
	query := `select back_key from users where id = :id`
	user := &core.User{}
	err := tx.Get(user, query, arg)
	if err != nil {
		return errors.NotFound("User not found")
	}

	// if user is manager:
	if len(user.BackKey) > 0 {
		return errors.BadRequest(errMessage)
	}
	return nil
}

func changeUserBlocked(userID uint, block bool, tx *db.Tx) error {
	query := `UPDATE users SET blocked = :blocked WHERE id = :id`

	arg := map[string]interface{}{
		"id":       userID,
		"blocked": block,
	}

	_ , err := tx.Exec(query, arg)
	return err
}

func CheckRootUser(rootPassword string, db *db.DB) (*core.User, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	user, err := checkRootUser(rootPassword, NewUserService(tx))
	if err != nil {
		if err := tx.Rollback(); err != nil {
		}
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return user, nil
}

func checkRootUser(rootPassword string, svc core.UserService) (*core.User, error) {
	user, err := svc.UserByUsername(core.RootUsername)
	if err != nil && err != core.UserNotFound {
		return nil, err
	}

	if user == nil {
		password, err := core.HashPassword(rootPassword)
		if err != nil {
			return nil, err
		}

		user = &core.User{
			Username: core.RootUsername,
			Password: password,
		}

		if err := svc.CreateUser(user); err != nil {
			return nil, err
		}
	} else {
		user.Password, err = core.HashPassword(rootPassword)
		if err != nil {
			return nil, err
		}

		if err := svc.UpdateUserPassword(user.ID, user.Password); err != nil {
			return nil, err
		}
	}

	return user, nil
}

func CheckManagers(managers map[string]string, db *db.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if err := checkManagers(managers, NewUserService(tx)); err != nil {
		if err := tx.Rollback(); err != nil {
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func checkManagers(managers map[string]string, svc core.UserService) error {
	users, err := svc.Managers()
	if err != nil {
		return err
	}
	for _, user := range users {
		if _, ok := managers[user.Username]; !ok && user.Level() == core.LevelManager {
			return fmt.Errorf("not found manager user: %s", user.Username)
		}
	}
	return nil
}
