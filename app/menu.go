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
		ctx.MainMenu = []MenuItem{NewMenuItem("Home", "/main/"),
			NewMenuItem("Network", "/net/"),
			NewMenuItem("Storage", "/storage/"),
			NewMenuItem("&nbsp;", "#"),
			NewMenuItem("MPD", "/mpd/"),
			NewMenuItem("&nbsp;", "#"),
			NewMenuItem("Utilities", "/utils/"),
			NewMenuItem("&nbsp;", "#"),
			NewMenuItem("Logout", "/auth/logoff")}
	} else {
		ctx.MainMenu = []MenuItem{NewMenuItem("Login", "/auth/login")}
	}
}
