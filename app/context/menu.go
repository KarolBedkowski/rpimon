package context

import (
	"container/list"
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

	MenuGenerator func(ctx *BasePageContext) (parentId string, menu *MenuItem)
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

func (i *MenuItem) AppendItemToParent(parentID string, item *MenuItem) (attached bool) {
	if i.ID == parentID {
		i.Submenu = append(i.Submenu, item)
		return true
	}
	if i.Submenu != nil {
		for _, subitem := range i.Submenu {
			if subitem.AppendItemToParent(parentID, item) {
				return true
			}
		}
	}
	return false
}

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

func (i *MenuItem) Sort() {
	if i.Submenu != nil {
		sort.Sort(subMenu(i.Submenu))
		for _, item := range i.Submenu {
			item.Sort()
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

// SetMainMenu - fill MainMenu in BasePageContext
func SetMainMenu(ctx *BasePageContext) {
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
	itemsLen := itemsWithoutParent.Len()
	for {
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
		mitem := &MenuItem{Title: "Other"}
		for e := itemsWithoutParent.Front(); e != nil; e = e.Next() {
			mitem.Submenu = append(mitem.Submenu, e.Value.(notAttachedItems).item)
		}
		ctx.MainMenu.AppendItemToParent("", mitem)
	}

	ctx.MainMenu.Sort()
}
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
