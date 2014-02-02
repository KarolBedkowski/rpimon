package database

import (
	"crypto/md5"
	"fmt"
	"io"
)

// User structure
type User struct {
	Login    string
	Password string
	Privs    []string
}

// GetUserByLogin - find user by login
func GetUserByLogin(login string) *User {
	for _, user := range database.Users {
		if user.Login == login {
			return &user
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
	hash := md5.New()
	io.WriteString(hash, candidatePassword)
	pass := fmt.Sprintf("%x", hash.Sum(nil))
	return user.Password == pass
}
