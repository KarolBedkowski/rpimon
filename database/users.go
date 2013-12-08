package database

type User struct {
	Login    string
	Password string
	Privs    []string
}

func GetUserByLogin(login string) *User {
	for _, user := range database.Users {
		if user.Login == login {
			return &user
		}
	}
	return nil
}
