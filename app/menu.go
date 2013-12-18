package app

// MenuItem - one position in menu
type MenuItem struct {
	Title   string
	Href    string
	Submenu []MenuItem
}

// NewMenuItem create new MenuItem structure
func NewMenuItem(title, href string) MenuItem {
	return MenuItem{Title: title, Href: href}
}

// SetMainMenu - fill MainMenu in BasePageContext
func SetMainMenu(ctx *BasePageContext, loggedUser bool) {
	if loggedUser {
		ctx.MainMenu = []MenuItem{NewMenuItem("Home", "/main/"),
			NewMenuItem("Network", "/net/"),
			NewMenuItem("Storage", "/storage/"),
			NewMenuItem("Logs", "/logs/"),
			NewMenuItem("Process", "/process/"),
			NewMenuItem("Users", "/users/"),
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
