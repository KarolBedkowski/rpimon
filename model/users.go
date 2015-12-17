package model

import (
	"bytes"
	"code.google.com/p/go.crypto/bcrypt"
	"encoding/gob"
	"errors"
	"io"
	l "k.prv/rpimon/logging"
)

var (
	// ErrUserExists - user already exists in system
	ErrUserExists = errors.New("user already exists")
	// ErrUserNotFound - user not exists in system
	ErrUserNotFound = errors.New("user not found")
)

type (
	User struct {
		Login    string
		Password string
		Privs    []string
	}
)

func GetUserByLogin(login string) (u *User) {
	v, err := db.db.Get(nil, login2key(login))
	if err != nil {
		l.Warn("globals.GetUser error: %s", err)
	}
	if v == nil {
		l.Debug("globals.GetUser user not found login=%s", login)
		return nil
	}
	return decodeUser(v)
}

func login2key(login string) []byte {
	return append(userPrefix, []byte(login)...)
}

func decodeUser(buff []byte) (u *User) {
	u = &User{}
	r := bytes.NewBuffer(buff)
	dec := gob.NewDecoder(r)
	err := dec.Decode(u)
	if err == nil {
		return u
	}
	l.Warn("globals.decodeUser decode error: %s", err)
	return nil
}

// AddUser into database; check is login unique
func AddUser(user *User) error {
	// check is login unique
	_, hit, _ := db.db.Seek(login2key(user.Login))
	if hit {
		return ErrUserExists
	}
	saveUser(user)
	return nil
}

func saveUser(u *User) (err error) {
	l.Info("globals.SaveUser %s", u)
	r := new(bytes.Buffer)
	enc := gob.NewEncoder(r)
	if err = enc.Encode(u); err != nil {
		l.Warn("SaveUser encode error: %s - %s", u, err)
		return
	}
	if err = db.db.Set(login2key(u.Login), r.Bytes()); err != nil {
		l.Warn("SaveUser set error %s: %s", u, err)
	}
	return
}

func GetAllUsers() (users []*User) {
	en, _, err := db.db.Seek(userPrefix)
	if err != nil {
		return
	}
	for {
		key, value, err := en.Next()
		if err == io.EOF || !bytes.HasPrefix(key, userPrefix) {
			break
		}
		if err == nil {
			users = append(users, decodeUser(value))
		} else {
			l.Error("GetUsers next error: %s", err)
		}
	}
	return
}

// UpdateUser in database.
func UpdateUser(user *User) error {
	return saveUser(user)
}

// DeleteUser from database by login
func DeleteUser(login string) error {
	return db.db.Delete(login2key(login))
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

// UpdatePassword for user
func (user *User) UpdatePassword(password string) {
	pass := CreatePassword(password)
	if pass != "" {
		user.Password = pass
	}
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
