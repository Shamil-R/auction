package core

import (
	"gitlab/nefco/auction/core/object"
	"gitlab/nefco/auction/errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	RootUserID   = 1
	RootUsername = "root"
	LevelRoot    = "root"
	LevelManager = "manager"
	LevelUser    = "user"
)

var (
	UserNotFound = errors.NotFound("User not found")
)

type UserInfo struct {
	EmployerName       string `json:"employer_name" db:"employer_name"`
	EmployerSurname    string `json:"employer_surname" db:"employer_surname"`
	EmployerPatronymic string `json:"employer_patronymic" db:"employer_patronymic"`
	CompanyName        string `json:"company_name" db:"company_name"`
	PhoneNumber        string `json:"phone_number" db:"phone_number"`
}

type User struct {
	UserInfo
	ID         uint              `json:"id" db:"id"`
	Username   string            `json:"username" db:"username" validate:"required"`
	Password   string            `json:"password,omitempty" db:"password" validate:"required"`
	Blocked    bool              `json:"blocked" db:"blocked"`
	ObjectType object.ObjectType `json:"object_type" db:"object_type" validate:"required,object_type"`
	BackKey    string            `json:"back_key,omitempty" db:"back_key" groups:"root"`
	Groups     []*Group          `json:"groups,omitempty"`
	lotsFilter LotsFilter
}

func (u *User) Level() string {
	if u.Username == RootUsername {
		return LevelRoot
	}
	if len(strings.TrimSpace(u.BackKey)) > 0 {
		return LevelManager
	}
	return LevelUser
}

func (u *User) Check(lot *Lot) bool {
	for _, group := range u.Groups {
		if group.Key == lot.GroupKey {
			return true
		}
	}
	return false
}

func (u *User) SetLotsFilter(f LotsFilter) {
	u.lotsFilter = f
}

func (u *User) CheckFilter(l *Lot) bool {
	f := u.lotsFilter
	if f != nil && ((f.LotID() > 0 && l.ID != f.LotID()) ||
		(f.ObjectID() > 0 && l.ObjectID != f.ObjectID()) ||
		(f.StartPrice() > 0 && l.BasePrice < f.StartPrice()) ||
		(f.EndPrice() > 0 && l.BasePrice > f.EndPrice()) ||
		!l.Object.Data.CheckFilter(f.Filter()) ||
		!l.Object.Type.FilterConfirm(l.Confirm, f.Filter()) ||
		((f.State() == FilterStateBooked || f.State() == FilterStateCompleted) &&
			l.BookedAt != nil && l.Bet != nil && l.Bet.UserID != u.ID)) {
		return false
	}
	return true
}

type UserGroup interface {
	GetUserID() uint
	GetGroupKey() string
}

type UserService interface {
	Users() ([]*User, error)
	Managers() ([]*User, error)
	User(id uint) (*User, error)
	UserByUsername(username string) (*User, error)
	CreateUser(user *User) error
	SaveUser(user *User) error
	UpdateUserPassword(userID uint, password string) error
	// DeleteUser(user *User) error
	AddGroup(ug UserGroup) error
	DeleteGroup(ug UserGroup) error
	GroupsByUser(userID uint) ([]*Group, error)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
