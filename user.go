package wilhelmiina

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	UUID      string `gorm:"primaryKey"`
	Username  string `gorm:"index:unique"`
	Firstname string
	Surname   string
	Password  string
	Role      Role
}

type UserData struct {
	UUID      string
	Username  string
	Firstname string
	Surname   string
}

type GuardianData struct {
	gorm.Model
	UUID       string
	GuardianOf string
}

type Role int

const Student Role = 0
const Guardian Role = 1
const Teacher Role = 2
const Moderator Role = 3
const Admin Role = 4

var ErrUserAlreadyExists = errors.New("username must be unique")

// Creates user and saves it to the database specified in the database argument. You should have migrated user schema to db already
func CreateUser(username string, Firstname string, Surname string, password string, role Role, database *gorm.DB) (User, error) {
	_, err := GetUserByUn(username, database)
	exists := err != ErrUserNotFound
	if exists {
		return User{}, ErrUserAlreadyExists
	}
	UUID := uuid.New()
	UUIDString := UUID.String()

	hashed, err := genHashString(password)
	if err != nil {
		return User{}, nil
	}
	u := User{
		UUID:      UUIDString,
		Username:  username,
		Firstname: Firstname,
		Surname:   Surname,
		Password:  hashed,
		Role:      role,
	}

	// Save user in database
	tx := database.Begin()
	database.Create(&u)
	err = tx.Commit().Error

	if err != nil {
		return User{}, err
	}
	return u, nil
}

var ErrUserNotFound = errors.New("user with that id not found")

func GetUser(UUID string, database *gorm.DB) (User, error) {
	var user User
	res := database.First(&user, "UUID = ?", UUID)
	found := res.RowsAffected != 0
	if !found {
		return User{}, ErrUserNotFound
	}
	return user, nil
}

//
func GetUserByUn(un string, db *gorm.DB) (User, error) {
	var user User
	res := db.First(&user, "username = ?", un)
	found := res.RowsAffected != 0
	if !found {
		return User{}, ErrUserNotFound
	}
	return user, nil
}

func ChangeUserNames(firstname string, lastname string, UUID string, db *gorm.DB) error {
	tx := db.Begin()
	tx.Model(User{}).Where("uuid = ?", UUID).
		Update("firstname", firstname).
		Update("lastname", lastname)

	err := tx.Commit().Error
	return err
}

func ChangePassword(newpass string, UUID string, db *gorm.DB) error {
	newHash, err := genHashString(newpass)
	if err != nil {
		return err
	}

	tx := db.Begin()
	tx.Model(User{}).Where("uuid = ?", UUID).Update("password", newHash)
	err = tx.Commit().Error
	return err
}

func DeleteUser(UUID string, db *gorm.DB) error {
	tx := db.Begin()
	tx.Where("uuid = ?").Delete(&User{})
	err := tx.Commit().Error
	return err
}

func (u *User) GetGroups(db *gorm.DB) ([]Group, error) {
	return GetUserGroups(u.UUID, db)
}

func (u *User) GetMessages(db *gorm.DB) ([]Message, error) {
	return GetMessagesForId(u.UUID, db)
}

func (u *User) SendMessage(toIDs []string, title string, content string, replyTo string, db *gorm.DB) (Message, error) {
	toIDs = append(toIDs, u.UUID)
	return SendMessage(u.UUID, toIDs, title, content, replyTo, db)
}

func (u *User) CheckPassword(password string) (bool, error) {
	return validatePassword(password, u.Password)
}

func (u *User) JoinGroup(groupID string, db *gorm.DB) (GroupReservation, error) {
	return CreateReservation(u.UUID, groupID, db)
}

func (u *User) UpdatePassword(newpass string, db *gorm.DB) error {
	return ChangePassword(newpass, u.UUID, db)
}

// ToData converts User database model to UserData because you probably don't want to give password hashes away in api replies
func (u *User) ToData() UserData {
	return UserData{
		UUID:      u.UUID,
		Username:  u.Username,
		Firstname: u.Firstname,
		Surname:   u.Surname,
	}
}
