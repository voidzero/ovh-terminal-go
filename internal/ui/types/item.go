// internal/ui/types/item.go
package types

import (
	"fmt"
	"io"
	"strings"

	"ovh-terminal/internal/ui/common"
	"ovh-terminal/internal/ui/styles"

	"github.com/charmbracelet/bubbles/list"
)

// Verify interface compliance at compile time
var _ common.MenuItem = ListItem{}

// ListItem represents a single item in the menu
type ListItem struct {
	text       string
	desc       string
	itemType   common.ItemType
	expanded   bool
	indent     int
	selectable bool
}

// MenuItemOption is a function type for applying options to a ListItem
type MenuItemOption func(*ListItem)

// Interface implementation
func (i ListItem) Title() string            { return i.text }
func (i ListItem) Description() string      { return i.desc }
func (i ListItem) FilterValue() string      { return i.text }
func (i ListItem) GetType() common.ItemType { return i.itemType }
func (i ListItem) IsExpanded() bool         { return i.expanded }
func (i ListItem) GetIndent() int           { return i.indent }
func (i ListItem) IsSelectable() bool       { return i.selectable }

// WithDesc sets the description
func WithDesc(desc string) MenuItemOption {
	return func(i *ListItem) {
		i.desc = desc
	}
}

// WithIndent sets the indentation level
func WithIndent(indent int) MenuItemOption {
	return func(i *ListItem) {
		i.indent = indent
	}
}

// WithExpanded sets the expanded state
func WithExpanded(expanded bool) MenuItemOption {
	return func(i *ListItem) {
		i.expanded = expanded
	}
}

// WithSelectable sets whether the item can be selected
func WithSelectable(selectable bool) MenuItemOption {
	return func(i *ListItem) {
		i.selectable = selectable
	}
}

// NewListItem creates a new ListItem with options
func NewListItem(text string, itemType common.ItemType, opts ...MenuItemOption) ListItem {
	item := ListItem{
		text:       text,
		itemType:   itemType,
		selectable: true, // default to selectable
	}

	for _, opt := range opts {
		opt(&item)
	}

	return item
}

// For backward compatibility
func (i ListItem) WithIndent(indent int) ListItem {
	newItem := i
	WithIndent(indent)(&newItem)
	return newItem
}

// For compatibility with list.Item interface
func (i ListItem) WithExpanded(expanded bool) list.Item {
	newItem := i
	newItem.expanded = expanded
	return newItem
}

// ItemDelegate handles the rendering of list items
type ItemDelegate struct {
	list.DefaultDelegate
}

// NewItemDelegate creates a new delegate with default styling
func NewItemDelegate() ItemDelegate {
	delegate := ItemDelegate{
		DefaultDelegate: list.DefaultDelegate{
			ShowDescription: false,
		},
	}

	delegate.Styles.SelectedTitle = styles.SelectedItemStyle
	delegate.Styles.SelectedDesc = styles.DimmedStyle
	delegate.Styles.NormalTitle = styles.NormalItemStyle
	delegate.Styles.NormalDesc = styles.DimmedStyle

	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		UnsetPadding().
		UnsetMargins()

	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		UnsetPadding().
		UnsetMargins()

	return delegate
}

// Render implements custom rendering for list items
func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	li, ok := item.(ListItem)
	if !ok {
		return
	}

	indent := strings.Repeat(" ", li.indent)
	var symbol string

	var prefix string
	switch li.itemType {
	case common.TypeTreeItem:
		prefix = "├─ "
	case common.TypeTreeLastItem:
		prefix = "└─ "
	case common.TypeHeader:
		if li.expanded {
			symbol = "[-] "
		} else {
			symbol = "[+] "
		}
	}

	title := indent + prefix + symbol + li.text
	style := styles.NormalItemStyle
	if index == m.Index() && li.selectable {
		style = styles.SelectedItemStyle
	}

	fmt.Fprint(w, style.Render(title))
}

// CreateBaseMenuItems returns the initial menu items
func CreateBaseMenuItems() []list.Item {
	return []list.Item{
		NewListItem("Account Information", common.TypeHeader,
			WithSelectable(true)),
		NewListItem("Bare Metal Cloud", common.TypeHeader,
			WithSelectable(true)),
		NewListItem("Web Cloud", common.TypeHeader,
			WithSelectable(true)),
		NewListItem("Exit", common.TypeNormal,
			WithDesc("Exit the application"),
			WithSelectable(true)),
	}
}
