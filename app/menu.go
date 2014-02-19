package app

// MenuItem - one position in menu
type MenuItem struct {
	Title   string
	Href    string
	ID      string
	Submenu []*MenuItem
	Icon    string
	Active  bool
}

// NewMenuItem create new MenuItem structure
func NewMenuItem(title, href string) *MenuItem {
	return &MenuItem{Title: title, Href: href, ID: href, Icon: "empty-icon"}
}

// NewMenuItemFromRoute create new menu item pointing to named route
func NewMenuItemFromRoute(title, routeName string, args ...string) *MenuItem {
	url := GetNamedURL(routeName, args...)
	return &MenuItem{Title: title, Href: url, ID: routeName, Icon: "empty-icon"}
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

// SetActve for menu item
func (item *MenuItem) SetActve(active bool) *MenuItem {
	item.Active = active
	return item
}

// AddChild append menu item as submenu item
func (item *MenuItem) AddChild(child *MenuItem) *MenuItem {
	item.Submenu = append(item.Submenu, child)
	return item
}

func (item *MenuItem) AttachSubmenu(parentID string, submenu []*MenuItem) (attached bool) {
	if item.ID == parentID {
		item.Submenu = append(item.Submenu, submenu...)
		return true
	}
	if item.Submenu != nil {
		for _, subitem := range item.Submenu {
			if subitem.AttachSubmenu(parentID, submenu) {
				return true
			}
		}
	}
	return false
}

func (item *MenuItem) SetActiveMenu(menuID string) (found bool) {
	if item.ID == menuID {
		item.Active = true
		return true
	}
	if item.Submenu != nil {
		for _, subitem := range item.Submenu {
			if subitem.SetActiveMenu(menuID) {
				item.Active = true
				return true
			}
		}
	}
	return false
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
				NewMenuItemFromRoute("Users", "users-index").SetID("users").SetIcon("glyphicon glyphicon-user")}
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
		// Tools
		toolsMenu := NewMenuItem("Tools", "").SetIcon("glyphicon glyphicon-briefcase").SetID("tools")
		if CheckPermission(ctx.CurrentUserPerms, "admin") {
			toolsMenu.Submenu = append(toolsMenu.Submenu,
				NewMenuItemFromRoute("Utilities", "utils-index").SetID("utils").SetIcon("glyphicon glyphicon-wrench"),
				NewMenuItem("-", ""))
		}
		if CheckPermission(ctx.CurrentUserPerms, "notepad") {
			toolsMenu.Submenu = append(toolsMenu.Submenu,
				NewMenuItemFromRoute("Notepad", "notepad-index").SetID("notepad-index").SetIcon("glyphicon glyphicon-paperclip"))
		}
		if toolsMenu.Submenu != nil {
			ctx.MainMenu = append(ctx.MainMenu, toolsMenu)
		}
	}
}

func AttachSubmenu(ctx *BasePageContext, parentID string, submenu []*MenuItem) {
	if ctx.MainMenu == nil {
		return
	}
	for _, subitem := range ctx.MainMenu {
		if subitem.AttachSubmenu(parentID, submenu) {
			return
		}
	}
}

// SetMenuActive add id  to menu active items
func MenuListSetMenuActive(id string, menu []*MenuItem) {
	for _, subitem := range menu {
		if subitem.SetActiveMenu(id) {
			break
		}
	}
}
