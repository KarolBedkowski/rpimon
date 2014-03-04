package cfg

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	l "k.prv/rpimon/helpers/logging"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
)

// User structure
type User struct {
	Login    string
	Password string
	Privs    []string
}

// UserDB - virtual database
type UserDB struct {
	Users []*User
}

var database UserDB

var AllPrivs = []string{
	"admin", "mpd", "files", "notepad",
}

var dbfilename string

// Init structures
func InitUsers(filename string, debug bool) {
	l.Info("UserDB.Init from: %s", filename)
	dbfilename = filename
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		l.Error("UserDB.Init Error:", err)
		createDummyDatabase()
		return
	}
	err = json.Unmarshal(file, &database)
	if err != nil {
		l.Error("UserDB.Init Error: %s", err.Error())
		createDummyDatabase()
	}
	l.Info("UserDB.Init loaded users: %d", len(database.Users))
	return
}

func usersSave() error {
	l.Printf("UserDB.usersSave %s\n", dbfilename)
	data, err := json.Marshal(database)
	if err != nil {
		l.Printf("UserDB.usersSave: error marshal configuration: %s\n", err)
		return err
	}
	return ioutil.WriteFile(dbfilename, data, 0)
}

func createDummyDatabase() {
	l.Info("Creating default user 'user', 'guest', 'admin'")
	//Create fake user
	database.Users = []*User{
		&User{
			Login:    "guest",
			Password: "",
			Privs:    nil,
		},
		&User{
			Login:    "user",
			Password: "",
			Privs:    []string{"mpd", "files"},
		},
		&User{
			Login:    "admin",
			Password: "",
			Privs:    []string{"admin", "mpd", "files"},
		},
	}
}
func GetUsers() []*User {
	return database.Users
}

// GetUserByLogin - find user by login
func GetUserByLogin(login string) *User {
	for _, user := range database.Users {
		if user.Login == login {
			return user
		}
	}
	return nil
}

// HasPermission check is user has given permission
func (user *User) HasPermission(permission string) bool {
	if permission == "" {
		return true
	}
	for _, perm := range user.Privs {
		if perm == permission {
			return true
		}
	}
	return false
}

// CheckPassword verify given password for user
func (user *User) CheckPassword(candidatePassword string) bool {
	l.Info("%#v %v", user, candidatePassword)
	if user.Password == "" {
		return candidatePassword == user.Login
	}
	pass := CreatePassword(candidatePassword)
	return user.Password == pass
}

func CreatePassword(password string) (encoded string) {
	if password == "" {
		return ""
	}
	hash := md5.New()
	io.WriteString(hash, password)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func AddUser(user *User) error {
	// check is login unique
	if GetUserByLogin(user.Login) != nil {
		return ErrUserExists
	}
	database.Users = append(database.Users, user)
	return usersSave()
}

func UpdateUser(user *User) error {
	for _, u := range database.Users {
		if u.Login == user.Login {
			u.Privs = user.Privs
			if user.Password != "" {
				u.Password = user.Password
			}
			return usersSave()
		}
	}
	return ErrUserNotFound
}

func DeleteUser(login string) error {
	if login == "admin" {
		return errors.New("can't remove admin")
	}
	for idx, u := range database.Users {
		if u.Login == login {
			database.Users = append(database.Users[:idx], database.Users[idx+1:]...)
			return nil
		}
	}
	return ErrUserNotFound
}
