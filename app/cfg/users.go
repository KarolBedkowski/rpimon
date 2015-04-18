package cfg

// User configuration
// Dummy json file database

import (
	"code.google.com/p/go.crypto/bcrypt"
	"encoding/json"
	"errors"
	"io/ioutil"
	"fmt"
	l "k.prv/rpimon/helpers/logging"
)

var (
	// ErrUserExists - user already exists in system
	ErrUserExists = errors.New("user already exists")
	// ErrUserNotFound - user not exists in system
	ErrUserNotFound = errors.New("user not found")
)

type (
	// User structure
	User struct {
		Login    string
		Password string
		Privs    []string
	}

	// UserDB - virtual database
	UserDB struct {
		// User login -> User
		Users map[string]*User
	}
)

var database UserDB

var dbfilename string

// InitUsers initialize users database
func InitUsers(filename string, debug bool) {
	l.Info("UserDB.Init from: %s", filename)
	dbfilename = filename
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		l.Error("UserDB.Init Error: " + err.Error())
		createDummyDatabase()
		saveUsers()
		return
	}
	err = json.Unmarshal(file, &database)
	if err != nil {
		l.Error("UserDB.Init Error: " + err.Error())
		createDummyDatabase()
		saveUsers()
	}
	l.Info("UserDB.Init loaded users: %d", len(database.Users))
	return
}

// save changes to database
func saveUsers() error {
	l.Printf("UserDB.saveUsers %s\n", dbfilename)
	data, err := json.Marshal(database)
	if err != nil {
		l.Error("UserDB.saveUsers: error marshal configuration: " + err.Error())
		return err
	}
	err = ioutil.WriteFile(dbfilename, data, 0600)
	if err != nil {
		l.Error(fmt.Sprintf("UserDB.saveUsers: error writing file %s: %s\nData: %v", dbfilename, err.Error(), data))
	}
	return err
}

func createDummyDatabase() {
	l.Info("Creating default user 'user', 'guest', 'admin'")
	//Create fake user
	database.Users = map[string]*User{
		"guest": {
			Login:    "guest",
			Password: CreatePassword("guest"),
			Privs:    nil,
		},
		"user": {
			Login:    "user",
			Password: CreatePassword("user"),
			Privs:    []string{"mpd", "files", "notepad"},
		},
		"admin": {
			Login:    "admin",
			Password: CreatePassword("admin"),
			Privs:    []string{"admin", "mpd", "files", "notepad"},
		},
	}
}

// GetAllUsers returns all known users as list
func GetAllUsers() (users []*User) {
	for _, u := range database.Users {
		users = append(users, u)
	}
	return
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

// CreatePassword create safe password from given string
func CreatePassword(password string) (encoded string) {
	if password == "" {
		return ""
	}

	data, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		l.Error("CreatePassword error: " + err.Error())
		return ""
	}
	return string(data)
}

// AddUser into database; check is login unique
func AddUser(user *User) error {
	// check is login unique
	if _, found := database.Users[user.Login]; found {
		return ErrUserExists
	}
	database.Users[user.Login] = user
	return saveUsers()
}

// UpdateUser in database.
func UpdateUser(user *User) error {
	if u, found := database.Users[user.Login]; found {
		u.Privs = user.Privs
		if user.Password != "" {
			u.Password = user.Password
		}
		return saveUsers()
	}
	return ErrUserNotFound
}

// DeleteUser from database by login
func DeleteUser(login string) error {
	if login == "admin" {
		return errors.New("can't remove admin")
	}
	if _, found := database.Users[login]; found {
		delete(database.Users, login)
		return nil
	}
	return ErrUserNotFound
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
		return candidatePassword == ""
	}
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(candidatePassword))
	if err != nil {
		l.Warn("CheckPassword for user %s error: %s", user.Login, err.Error())
		return false
	}
	return true
}

// UpdatePassword for user
func (user *User) UpdatePassword(password string) {
	pass := CreatePassword(password)
	if pass != "" {
		user.Password = pass
	}
}
