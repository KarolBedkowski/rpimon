package context

import (
	"container/list"
	l "k.prv/rpimon/logging"
	"sort"
)

type (
	// MenuItem - one position in menu
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

	// MenuGenerator function generate menu items in given context
	MenuGenerator func(ctx *BaseCtx) (parentId string, menu *MenuItem)
)

// NewMenuItem create new MenuItem structure
func NewMenuItem(title, href string) *MenuItem {
	return &MenuItem{Title: title, Href: href, ID: href, Icon: "empty-icon"}
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

// SetSortOrder for menu item
func (item *MenuItem) SetSortOrder(sortOrder int) *MenuItem {
	item.SortOrder = sortOrder
	return item
}

// AddChild append menu item as submenu item
func (item *MenuItem) AddChild(child ...*MenuItem) *MenuItem {
	item.Submenu = append(item.Submenu, child...)
	return item
}

//AppendItemToParent look for given menu by id in menu and append item as sub menu
func (item *MenuItem) AppendItemToParent(parentID string, newitem *MenuItem) (attached bool) {
	if item.ID == parentID {
		item.Submenu = append(item.Submenu, newitem)
		return true
	}
	if item.Submenu != nil {
		for _, subitem := range item.Submenu {
			if subitem.AppendItemToParent(parentID, newitem) {
				return true
			}
		}
	}
	return false
}

// SetActiveMenuItem find menu item by id and set it active; also update all parents
func (item *MenuItem) SetActiveMenuItem(menuID string) (found bool) {
	if item.ID == menuID {
		item.Active = true
		return true
	}
	if item.Submenu != nil {
		for _, subitem := range item.Submenu {
			if subitem.SetActiveMenuItem(menuID) {
				item.Active = true
				return true
			}
		}
	}
	return false
}

// Sort menu item and all submenu
func (item *MenuItem) Sort() {
	if item.Submenu != nil {
		sort.Sort(subMenu(item.Submenu))
		for _, sitem := range item.Submenu {
			sitem.Sort()
		}
	}
}

// SORTING MENU ITEMS

type subMenu []*MenuItem

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

// SetMainMenu - fill MainMenu in BaseCtx
func SetMainMenu(ctx *BaseCtx) {
	ctx.MainMenu = &MenuItem{}
	itemsWithoutParent := list.New()
	for _, module := range GetModules() {
		if module.Enabled() && module.GetMenu != nil {
			parent, mitem := module.GetMenu(ctx)
			if mitem != nil {
				if !ctx.MainMenu.AppendItemToParent(parent, mitem) {
					itemsWithoutParent.PushBack(notAttachedItems{parent, mitem})
				}
			}
		}
	}
	for {
		itemsLen := itemsWithoutParent.Len()
		if itemsWithoutParent.Len() == 0 {
			break
		}
		var next *list.Element
		for e := itemsWithoutParent.Front(); e != nil; e = next {
			next = e.Next()
			nai := e.Value.(notAttachedItems)
			if ctx.MainMenu.AppendItemToParent(nai.parent, nai.item) {
				itemsWithoutParent.Remove(e)
			}
		}
		if itemsLen == itemsWithoutParent.Len() {
			// items list not changed
			break
		}
	}
	if itemsWithoutParent.Len() > 0 {
		l.Warn("Items without parent len=%d", itemsWithoutParent.Len())
		mitem := &MenuItem{Title: "Other"}
		mitem.SetSortOrder(998)
		for e := itemsWithoutParent.Front(); e != nil; e = e.Next() {
			item := e.Value.(notAttachedItems).item
			l.Debug("Item without parent %#v", item)
			mitem.Submenu = append(mitem.Submenu, item)
		}
		ctx.MainMenu.AppendItemToParent("", mitem)
	}

	ctx.MainMenu.Sort()
}

// CheckPermission check if required permission is in the list.
func CheckPermission(userPermissions []string, required string) bool {
	if required == "" {
		return true
	}
	for _, p := range userPermissions {
		if p == required {
			return true
		}
	}
	return false
}
