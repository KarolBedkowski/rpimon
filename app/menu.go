package app

type MenuItem struct {
	Title   string
	Href    string
	Submenu []MenuItem
}

func NewMenuItem(title, href string) MenuItem {
	return MenuItem{Title: title, Href: href}
}

func SetMainMenu(ctx *BasePageContext, loggedUser bool) {
	if loggedUser {
		ctx.MainMenu = []MenuItem{NewMenuItem("Network", "/net/"),
			NewMenuItem("Storage", "/storage/"),
			NewMenuItem("&nbsp;", "#"),
			NewMenuItem("Logout", "/auth/logoff")}
	} else {
		ctx.MainMenu = []MenuItem{NewMenuItem("Login", "/auth/login")}
	}
}
