// internal/ui/types/item.go
package types

import (
	"fmt"
	"io"
	"strings"

	"ovh-terminal/internal/ui/common"
	"ovh-terminal/internal/ui/styles"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/charmbracelet/bubbles/list"
)

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

// Verify interface compliance at compile time
var (
	_ list.Item       = (*ListItem)(nil)
	_ common.MenuItem = (*ListItem)(nil)
)

// Interface implementation
func (i *ListItem) Title() string       { return i.text }
func (i *ListItem) Description() string { return i.desc }
func (i *ListItem) FilterValue() string { return i.text }

// MenuItem interface implementation
func (i *ListItem) GetType() common.ItemType { return i.itemType }
func (i *ListItem) IsExpanded() bool         { return i.expanded }
func (i *ListItem) GetIndent() int           { return i.indent }
func (i *ListItem) IsSelectable() bool       { return i.selectable }
func (i *ListItem) WithExpanded(expanded bool) list.Item {
	newItem := *i
	newItem.expanded = expanded
	return &newItem
}

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
func NewListItem(text string, itemType common.ItemType, opts ...MenuItemOption) *ListItem {
	item := &ListItem{
		text:       text,
		itemType:   itemType,
		selectable: true, // default to selectable
	}

	for _, opt := range opts {
		opt(item)
	}

	return item
}

// DefaultDelegate represents our custom item delegate
type DefaultDelegate struct {
	ShowDescription bool
	Styles          *list.DefaultItemStyles
}

// Height returns the height of the delegate
func (d DefaultDelegate) Height() int {
	return 1
}

// Spacing returns the spacing of the delegate
func (d DefaultDelegate) Spacing() int {
	return 0
}

// Update handles the update of the delegate
func (d DefaultDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

// Render implements custom rendering for list items
func (d DefaultDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	listItem, ok := item.(*ListItem)
	if !ok {
		return
	}

	// Helper to check if item has siblings after it
	hasNextSibling := func(idx int, indent int) bool {
		for i := idx + 1; i < len(m.Items()); i++ {
			if next, ok := m.Items()[i].(*ListItem); ok {
				if next.GetIndent() < indent {
					return false
				}
				if next.GetIndent() == indent {
					return true
				}
			}
		}
		return false
	}

	// Get parent continuation markers
	getParentMarkers := func(idx int, indent int) []bool {
		markers := make([]bool, indent)
		currentIndent := indent - 1
		for i := idx - 1; i >= 0 && currentIndent >= 0; i-- {
			if prev, ok := m.Items()[i].(*ListItem); ok {
				if prev.GetIndent() == currentIndent {
					markers[currentIndent] = hasNextSibling(i, currentIndent)
					currentIndent--
				}
			}
		}
		return markers
	}

	// Build the item's prefix
	var prefixParts []string

	// Handle indentation and tree structure
	if listItem.GetIndent() > 0 {
		// Get parent continuation markers
		markers := getParentMarkers(index, listItem.GetIndent())

		// Add spacing and vertical lines for each level
		for i := 0; i < listItem.GetIndent()-1; i++ {
			if markers[i] {
				prefixParts = append(prefixParts, "│  ")
			} else {
				prefixParts = append(prefixParts, "   ")
			}
		}

		// Add the appropriate connector
		if hasNextSibling(index, listItem.GetIndent()) {
			prefixParts = append(prefixParts, "├─")
		} else {
			prefixParts = append(prefixParts, "└─")
		}
	}

	// Ensure proper spacing with a leading space for indented items
	prefix := ""
	if len(prefixParts) > 0 {
		prefix = " " + strings.Join(prefixParts, "")
	}

	// Add expansion/collapse indicators for headers
	if listItem.itemType == common.TypeHeader {
		if listItem.expanded {
			prefix += "[-] "
		} else {
			prefix += "[+] "
		}
	} else if listItem.GetIndent() > 0 {
		prefix += " "
	}

	// Build the complete title
	completeTitle := prefix + listItem.text

	// Apply styling based on selection
	style := styles.NormalItemStyle
	if index == m.Index() && listItem.selectable {
		style = styles.SelectedItemStyle
	}

	fmt.Fprint(w, style.Render(completeTitle))
}

// NewDefaultDelegate creates a new delegate with default styling
func NewDefaultDelegate() DefaultDelegate {
	delegate := DefaultDelegate{
		ShowDescription: false,
		Styles: &list.DefaultItemStyles{
			NormalTitle:   styles.NormalItemStyle,
			SelectedTitle: styles.SelectedItemStyle,
			DimmedTitle:   styles.DimmedStyle,
			NormalDesc:    styles.DimmedStyle,
			SelectedDesc:  styles.DimmedStyle,
			DimmedDesc:    styles.DimmedStyle,
		},
	}

	return delegate
}

// CreateBaseMenuItems returns the initial menu items
func CreateBaseMenuItems() []list.Item {
	items := []list.Item{
		NewListItem("Account Information", common.TypeHeader),
		NewListItem("Bare Metal Cloud", common.TypeHeader),
		NewListItem("Web Cloud", common.TypeHeader),
		NewListItem("Exit", common.TypeNormal,
			WithDesc("Exit the application")),
	}
	return items
}

