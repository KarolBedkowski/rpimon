package app

import "k.prv/rpimon/app/context"

// NewMenuItemFromRoute create new menu item from named route
func NewMenuItemFromRoute(name string, route string, args ...string) *context.MenuItem {
	return context.NewMenuItem(name, GetNamedURL(route, args...)).SetID(route)
}
