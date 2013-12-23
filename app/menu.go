package app

// MenuItem - one position in menu
type MenuItem struct {
	Title   string
	Href    string
	ID      string
	Submenu []MenuItem
}

// NewMenuItem create new MenuItem structure
func NewMenuItem(title, href string) MenuItem {
	return MenuItem{Title: title, Href: href, ID: href}
}

func NewMenuItemFromRoute(title, name string, pairs ...string) MenuItem {
	url := GetNamedURL(name, pairs...)
	return MenuItem{Title: title, Href: url, ID: name}
}

func (item MenuItem) SetID(ID string) MenuItem {
	item.ID = ID
	return item
}

func (item MenuItem) AddQuery(query string) MenuItem {
	item.Href += query
	return item
}

// SetMainMenu - fill MainMenu in BasePageContext
func SetMainMenu(ctx *BasePageContext) {
	if ctx.CurrentUser != "" {
		user := GetUser(ctx.CurrentUser)
		ctx.MainMenu = []MenuItem{NewMenuItemFromRoute("Home", "main-index").SetID("main")}
		if user.HasPermission("admin") {
			ctx.MainMenu = append(ctx.MainMenu,
				NewMenuItemFromRoute("System", "main-system").SetID("system"),
				NewMenuItemFromRoute("Network", "net-index").SetID("net"),
				NewMenuItemFromRoute("Storage", "storage-index").SetID("storage"),
				NewMenuItemFromRoute("Logs", "logs-index").SetID("logs"),
				NewMenuItemFromRoute("Process", "process-index").SetID("process"),
				NewMenuItemFromRoute("Users", "users-index").SetID("users"),
				NewMenuItemFromRoute("Utilities", "utils-index").SetID("utils"))
		}
		if user.HasPermission("mpd") {
			ctx.MainMenu = append(ctx.MainMenu,
				NewMenuItem("&nbsp;", "#"),
				NewMenuItemFromRoute("MPD", "mpd-index").SetID("mpd"))
		}
		if user.HasPermission("files") {
			ctx.MainMenu = append(ctx.MainMenu,
				NewMenuItem("&nbsp;", "#"),
				NewMenuItemFromRoute("Files", "files-index").SetID("files"))
		}
		ctx.MainMenu = append(ctx.MainMenu,
			NewMenuItem("&nbsp;", "#"),
			NewMenuItemFromRoute("Logout", "auth-logoff").SetID("auth-logoff"))
		return
	}
	ctx.MainMenu = []MenuItem{NewMenuItemFromRoute("Login", "auth-login").SetID("auth-login")}
}
