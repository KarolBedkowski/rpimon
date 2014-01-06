package app

// MenuItem - one position in menu
type MenuItem struct {
	Title   string
	Href    string
	ID      string
	Submenu []*MenuItem
}

// NewMenuItem create new MenuItem structure
func NewMenuItem(title, href string) *MenuItem {
	return &MenuItem{Title: title, Href: href, ID: href}
}

// NewMenuItemFromRoute create new menu item pointing to named route
func NewMenuItemFromRoute(title, routeName string, args ...string) *MenuItem {
	url := GetNamedURL(routeName, args...)
	return &MenuItem{Title: title, Href: url, ID: routeName}
}

// SetID for menu item
func (item *MenuItem) SetID(ID string) *MenuItem {
	item.ID = ID
	return item
}

// AddQuery to menu item href
func (item *MenuItem) AddQuery(query string) *MenuItem {
	item.Href += query
	return item
}

// SetMainMenu - fill MainMenu in BasePageContext
func SetMainMenu(ctx *BasePageContext) {
	if ctx.CurrentUser != "" {
		user := GetUser(ctx.CurrentUser)
		ctx.MainMenu = []*MenuItem{NewMenuItemFromRoute("Home", "main-index").SetID("main")}
		if user.HasPermission("admin") {
			sysMI := NewMenuItem("System", "")
			sysMI.Submenu = []*MenuItem{
				NewMenuItemFromRoute("Live view", "main-system").SetID("system"),
				NewMenuItemFromRoute("Network", "net-index").SetID("net"),
				NewMenuItemFromRoute("Storage", "storage-index").SetID("storage"),
				NewMenuItemFromRoute("Logs", "logs-index").SetID("logs"),
				NewMenuItemFromRoute("Process", "process-index").SetID("process"),
				NewMenuItemFromRoute("Users", "users-index").SetID("users"),
				NewMenuItemFromRoute("Utilities", "utils-index").SetID("utils")}
			ctx.MainMenu = append(ctx.MainMenu, sysMI)
		}
		if user.HasPermission("mpd") {
			ctx.MainMenu = append(ctx.MainMenu,
				NewMenuItemFromRoute("MPD", "mpd-index").SetID("mpd"))
		}
		if user.HasPermission("files") {
			ctx.MainMenu = append(ctx.MainMenu,
				NewMenuItemFromRoute("Files", "files-index").SetID("files"))
		}
	}
}
