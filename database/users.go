package database

import (
	"../helpers"
	l "../helpers/logging"
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
	Code string
}

type ProfilePrivilage struct {
	PrivilageId int64
	ProfileId   int64
}

type UserProfile struct {
	UserId    int64
	ProfileId int64
}

func GetUserByLogin(login string) (user *User) {
	user = new(User)
	err := Database.SelectOne(user, "select * from users where login=?", login)
	if err != nil {
		return nil
	}
	return
}

func GetUserById(userId int64) (user *User) {
	obj, _ := Database.Get(User{}, userId)
	if obj == nil {
		return nil
	}
	return obj.(*User)
}

func (user *User) Save() {
	if user.Id > 0 {
		l.Info("Update user ", user.Id)
		cnt, err := Database.Update(user)
		helpers.CheckErr(err, "Update user failed")
		if cnt > 0 {
			return
		}
		l.Warn("Update missing user")
	}
	err := Database.Insert(user)
	helpers.CheckErr(err, "Update user failed")
}

func UsersList() []User {
	var users []User
	_, err := Database.Select(&users, "select * from users")
	helpers.CheckErr(err, "User list error")
	return users
}

func (priv *Privilage) Save() {
	if priv.Id > 0 {
		l.Info("Update priv ", priv.Id)
		cnt, err := Database.Update(priv)
		helpers.CheckErr(err, "Update priv failed")
		if cnt > 0 {
			return
		}
		l.Warn("Update missing priv")
	}
	err := Database.Insert(priv)
	helpers.CheckErr(err, "Update priv failed")
}

func GetPrivilage(privId int64) *Privilage {
	obj, _ := Database.Get(Privilage{}, privId)
	if obj == nil {
		return nil
	}
	return obj.(*Privilage)
}

func PrivilagesList() []Privilage {
	var privs []Privilage
	_, err := Database.Select(&privs, "select * from privilages")
	helpers.CheckErr(err, "Priviages list error")
	return privs
}

func GetUserPrivilages(userId int64) []Privilage {
	var privs []Privilage
	_, err := Database.Select(&privs,
		"select p.* "+
			"from privilages p, profile_privilages pp, user_profile up "+
			"where p.id = pp.PrivilageId and pp.ProfileId = up.ProfileId "+
			" and up.UserId = ?", userId)
	helpers.CheckErr(err, "User list error")
	return privs
}

func PrivilagesToStr(privs []Privilage) []string {
	res := make([]string, len(privs))
	for pos, prv := range privs {
		res[pos] = prv.Code
	}
	return res
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
	count, err = Database.SelectInt("select count(*) from privilages")
	helpers.CheckErr(err, "BootstrapUsers Count privilages failed")
	if count == 0 {
		admin := &Privilage{
			Name: "Admin",
			Code: "admin"}
		admin.Save()
		user := &Privilage{
			Name: "User",
			Code: "user"}
		user.Save()
	}
}
