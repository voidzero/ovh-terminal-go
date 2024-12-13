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
	text       string          // The text to display
	desc       string          // Description or additional info
	itemType   common.ItemType // Type of the item
	expanded   bool            // Whether a header is expanded
	indent     int             // Indentation level
	selectable bool            // Whether the item can be selected
}

// Title implements list.Item interface
func (i ListItem) Title() string       { return i.text }
func (i ListItem) Description() string { return i.desc }
func (i ListItem) FilterValue() string { return i.text }

// GetType returns the item's type
func (i ListItem) GetType() common.ItemType { return i.itemType }

// IsExpanded returns whether the item is expanded
func (i ListItem) IsExpanded() bool { return i.expanded }

// GetIndent returns the item's indentation level
func (i ListItem) GetIndent() int { return i.indent }

// IsSelectable returns whether the item can be selected
func (i ListItem) IsSelectable() bool { return i.selectable }

// NewListItem creates a new ListItem with basic initialization
func NewListItem(text string, itemType common.ItemType, desc string) ListItem {
	return ListItem{
		text:       text,
		desc:       desc,
		itemType:   itemType,
		expanded:   false,
		indent:     0,
		selectable: true,
	}
}

// Builder methods for fluent interface

// WithDesc sets the description and returns the modified item
func (i ListItem) WithDesc(desc string) ListItem {
	i.desc = desc
	return i
}

// WithIndent sets the indentation level and returns the modified item
func (i ListItem) WithIndent(indent int) ListItem {
	i.indent = indent
	return i
}

// WithExpanded sets the expanded state and returns the modified item
func (i ListItem) WithExpanded(expanded bool) list.Item {
	i.expanded = expanded
	return i
}

// WithSelectable sets whether the item can be selected and returns the modified item
func (i ListItem) WithSelectable(selectable bool) ListItem {
	i.selectable = selectable
	return i
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

	// Style the list items
	delegate.Styles.SelectedTitle = styles.SelectedItemStyle
	delegate.Styles.SelectedDesc = styles.DimmedStyle
	delegate.Styles.NormalTitle = styles.NormalItemStyle
	delegate.Styles.NormalDesc = styles.DimmedStyle

	// Remove padding and margins
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

	// Tree structure symbols
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
		NewListItem("Account Information", common.TypeHeader, "").
			WithSelectable(true),
		NewListItem("Bare Metal Cloud", common.TypeHeader, "").
			WithSelectable(true),
		NewListItem("Web Cloud", common.TypeHeader, "").
			WithSelectable(true),
		NewListItem("Exit", common.TypeNormal, "Exit the application").
			WithSelectable(true),
	}
}

