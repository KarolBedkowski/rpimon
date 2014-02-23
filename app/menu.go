package app

import (
	"container/list"
	"sort"
)

// MenuItem - one position in menu
type (
	MenuItem struct {
		Title     string
		Href      string
		ID        string
		Submenu   []*MenuItem
		Icon      string
		Active    bool
		SortOrder int
		// RequredPrivilages as [[priv and priv ....] or [ priv ...]]
		RequredPrivilages [][]string
	}
	subMenu []*MenuItem
)

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
func (item *MenuItem) AddChild(child ...*MenuItem) *MenuItem {
	item.Submenu = append(item.Submenu, child...)
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

func (i *MenuItem) AppendItem(parentID string, item *MenuItem) (attached bool) {
	if i.ID == parentID {
		i.Submenu = append(i.Submenu, item)
		return true
	}
	if i.Submenu != nil {
		for _, subitem := range i.Submenu {
			if subitem.AppendItem(parentID, item) {
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

func (i *MenuItem) Sort() {
	if i.Submenu != nil {
		sort.Sort(subMenu(i.Submenu))
		for _, item := range i.Submenu {
			item.Sort()
		}
	}
}

func (s subMenu) Len() int      { return len(s) }
func (s subMenu) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s subMenu) Less(i, j int) bool {
	if s[i].SortOrder == s[j].SortOrder {
		return s[i].Title < s[j].Title
	}
	return s[i].SortOrder < s[j].SortOrder
}

type notAttachedItems struct {
	parent string
	item   *MenuItem
}

// SetMainMenu - fill MainMenu in BasePageContext
func SetMainMenu(ctx *BasePageContext) {
	ctx.MainMenu = &MenuItem{}
	itemsWithoutParent := list.New()
	for _, item := range ModulesMenuItems {
		parent, mitem := item(ctx)
		if mitem != nil {
			if !ctx.MainMenu.AppendItem(parent, mitem) {
				itemsWithoutParent.PushBack(notAttachedItems{parent, mitem})
			}
		}
	}
	itemsLen := itemsWithoutParent.Len()
	for {
		if itemsWithoutParent.Len() == 0 {
			break
		}
		var next *list.Element
		for e := itemsWithoutParent.Front(); e != nil; e = next {
			next = e.Next()
			nai := e.Value.(notAttachedItems)
			if ctx.MainMenu.AppendItem(nai.parent, nai.item) {
				itemsWithoutParent.Remove(e)
			}
		}
		if itemsLen == itemsWithoutParent.Len() {
			// items list not changed
			break
		}
	}
	if itemsWithoutParent.Len() > 0 {
		mitem := &MenuItem{Title: "Other"}
		for e := itemsWithoutParent.Front(); e != nil; e = e.Next() {
			mitem.Submenu = append(mitem.Submenu, e.Value.(notAttachedItems).item)
		}
		ctx.MainMenu.AppendItem("", mitem)
	}

}

func AttachSubmenu(ctx *BasePageContext, parentID string, submenu []*MenuItem) {
	if ctx.MainMenu == nil {
		return
	}
	ctx.MainMenu.AttachSubmenu(parentID, submenu)
}

// SetMenuActive add id  to menu active items
func MenuListSetMenuActive(id string, menu []*MenuItem) {
	for _, subitem := range menu {
		if subitem.SetActiveMenu(id) {
			break
		}
	}
	ctx.MainMenu.Sort()
}

type GetMenuFunc func(ctx *BasePageContext) (parentId string, menu *MenuItem)

// TODO: przerobic
var ModulesMenuItems []GetMenuFunc

func RegisterMenuProvider(f GetMenuFunc) {
	ModulesMenuItems = append(ModulesMenuItems, f)
}
