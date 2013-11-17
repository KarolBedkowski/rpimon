package database

import (
	"../helpers"
	"log"
)

type User struct {
	Id       int64
	Name     string
	Password string
	Login    string
}

type Profile struct {
	Id   int64
	Name string
}

type Privilage struct {
	Id   int64
	Name string
}

func GetUserByLogin(login string) (user *User) {
	user = new(User)
	err := Database.SelectOne(user, "select * from users where login=?", login)
	if err != nil {
		user = nil
	}
	return
}

func GetUserById(userId int64) (user *User) {
	user = new(User)
	err := Database.SelectOne(user, "select * from users where id=?", userId)
	if err != nil {
		user = nil
	}
	return
}

func (user *User) Save() {
	if user.Id > 0 {
		log.Print("Update user ", user.Id)
		cnt, err := Database.Update(user)
		helpers.CheckErr(err, "Update user failed")
		if cnt == 0 {
			log.Fatal("Update missing user")
		}
	} else {
		err := Database.Insert(user)
		helpers.CheckErr(err, "Update user failed")
	}
}

func UsersList() []User {
	var users []User
	_, err := Database.Select(&users, "select * from users")
	helpers.CheckErr(err, "User list error")
	return users
}

func BootstrapUsers() {
	count, err := Database.SelectInt("select count(*) from users")
	helpers.CheckErr(err, "BootstrapUsers Count users failed")
	if count == 0 {
		admin := &User{
			Login:    "admin",
			Name:     "Administrator",
			Password: helpers.CreatePassword("admin")}
		admin.Save()
	}
}
