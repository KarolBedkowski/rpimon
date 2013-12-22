package database

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
