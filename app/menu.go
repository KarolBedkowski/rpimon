package app

// MenuItem - one position in menu
type MenuItem struct {
	Title   string
	Href    string
	ID      string
	Submenu []*MenuItem
	Icon    string
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

// SetIcon for menu item
func (item *MenuItem) SetIcon(icon string) *MenuItem {
	item.Icon = icon
	return item
}

// AddChild append menu item as submenu item
func (item *MenuItem) AddChild(child *MenuItem) *MenuItem {
	item.Submenu = append(item.Submenu, child)
	return item
}

// SetMainMenu - fill MainMenu in BasePageContext
func SetMainMenu(ctx *BasePageContext) {
	if ctx.CurrentUser != "" {
		ctx.MainMenu = []*MenuItem{NewMenuItemFromRoute("Home", "main-index").SetID("main").SetIcon("glyphicon glyphicon-home")}
		if CheckPermission(ctx.CurrentUserPerms, "admin") {
			sysMI := NewMenuItem("System", "").SetIcon("glyphicon glyphicon-wrench").SetID("system")
			sysMI.Submenu = []*MenuItem{
				NewMenuItemFromRoute("Live view", "main-system").SetID("system-live").SetIcon("glyphicon glyphicon-dashboard"),
				NewMenuItem("-", ""),
				NewMenuItemFromRoute("Network", "net-index").SetID("net").SetIcon("glyphicon glyphicon-transfer"),
				NewMenuItemFromRoute("Storage", "storage-index").SetID("storage").SetIcon("glyphicon glyphicon-hdd"),
				NewMenuItemFromRoute("Logs", "logs-index").SetID("logs").SetIcon("glyphicon glyphicon-eye-open"),
				NewMenuItemFromRoute("Process", "process-index").SetID("process").SetIcon("glyphicon glyphicon-cog"),
				NewMenuItemFromRoute("Users", "users-index").SetID("users").SetIcon("glyphicon glyphicon-user"),
				NewMenuItem("-", ""),
				NewMenuItemFromRoute("Utilities", "utils-index").SetID("utils").SetIcon("glyphicon glyphicon-wrench")}
			ctx.MainMenu = append(ctx.MainMenu, sysMI)
		}
		if CheckPermission(ctx.CurrentUserPerms, "mpd") {
			ctx.MainMenu = append(ctx.MainMenu,
				NewMenuItemFromRoute("MPD", "mpd-index").SetID("mpd").SetIcon("glyphicon glyphicon-music"))
		}
		if CheckPermission(ctx.CurrentUserPerms, "files") {
			ctx.MainMenu = append(ctx.MainMenu,
				NewMenuItemFromRoute("Files", "files-index").SetID("files").SetIcon("glyphicon glyphicon-hdd"))
		}

	}
}
