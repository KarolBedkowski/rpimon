package users

import (
	"../../app"
	"../../database"
	l "../../helpers/logging"
	"../../security"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type PrivsPageCtx struct {
	*app.BasePageContext
	Privilages []database.Privilage
}

func newPrivsPageCtx(w http.ResponseWriter, r *http.Request,
	privs []database.Privilage) *PrivsPageCtx {
	return &PrivsPageCtx{app.NewBasePageContext("Privilage", w, r), privs}
}

func getPrivsHandler(w http.ResponseWriter, r *http.Request) {
	if security.GetLoggedUser(w, r, true) == nil {
		return
	}
	data := newPrivsPageCtx(w, r, database.PrivilagesList())
	app.RenderTemplate(w, data, "base", "base.tmpl", "users/privs.tmpl",
		"flash.tmpl")
}

type PrivForm struct {
	database.Privilage
}

type EditPrivPageCtx struct {
	*app.BasePageContext
	*PrivForm
	Message string
}

func (form *PrivForm) validate() (err string) {
	err = ""
	if form.Code == "" {
		err = "Missing code"
	}
	if form.Name == "" {
		err = "Missing name"
		return
	}
	return
}

func newEditPrivPageCtx(w http.ResponseWriter, r *http.Request, msg string) *EditPrivPageCtx {
	return &EditPrivPageCtx{app.NewBasePageContext("Privilage", w, r),
		&PrivForm{}, msg}
}

func getEditPrivPageHandler(w http.ResponseWriter, r *http.Request) {
	if security.GetLoggedUser(w, r, true) == nil {
		return
	}
	editPage := newEditPrivPageCtx(w, r, "")
	vars := mux.Vars(r)
	privId, ok := vars["id"]
	if ok && privId != "" {
		privIdI, _ := strconv.ParseInt(privId, 10, 64)
		priv := database.GetPrivilage(privIdI)
		if priv != nil {
			editPage.Privilage = *priv
		}
	}
	app.RenderTemplate(w, editPage, "base", "base.tmpl", "users/priv_edit.tmpl", "flash.tmpl")
	return
}

func saveEditPrivPageHangler(w http.ResponseWriter, r *http.Request) {
	if security.GetLoggedUser(w, r, true) == nil {
		return
	}
	r.ParseForm()
	editPage := newEditPrivPageCtx(w, r, "")
	err := decoder.Decode(editPage, r.Form)
	if err != nil {
		l.Warn("Decoding form error", err)
	}
	msg := editPage.validate()
	if msg != "" {
		editPage.Message = msg
		app.RenderTemplate(w, editPage, "base", "base.tmpl", "users/priv_edit.tmpl", "flash.tmpl")
		return
	}
	var priv *database.Privilage
	if editPage.Id > 0 {
		priv = database.GetPrivilage(editPage.Id)
	} else {
		priv = new(database.Privilage)
	}
	priv.Name = editPage.Name
	priv.Code = editPage.Code
	priv.Save()
	url, _ := subRouter.Get("privs-list").URL()
	editPage.AddFlashMessage("Privilage Saved")
	editPage.SessionSave()
	http.Redirect(w, r, url.String(), http.StatusFound)
}
