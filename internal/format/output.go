// internal/format/output.go
package format

import (
	"fmt"
	"strings"
)

// Section represents a group of related data
type Section struct {
	Title   string
	Content [][]string // Rows of [key, value]
}

// OutputFormatter handles formatted text output
type OutputFormatter struct {
	sections []*Section
}

// NewOutputFormatter creates a new formatter instance
func NewOutputFormatter() *OutputFormatter {
	return &OutputFormatter{
		sections: make([]*Section, 0),
	}
}

// AddSection adds a new section to the output
func (f *OutputFormatter) AddSection(title string) *Section {
	section := &Section{
		Title:   title,
		Content: make([][]string, 0),
	}
	f.sections = append(f.sections, section)
	return section
}

// AddField adds a field to a section
func (s *Section) AddField(key, value string) {
	if value != "" {
		s.Content = append(s.Content, []string{key, value})
	}
}

// String formats the entire output
func (f *OutputFormatter) String() string {
	var output strings.Builder

	for i, section := range f.sections {
		if i > 0 {
			output.WriteString("\n")
		}

		// Write section title
		output.WriteString(fmt.Sprintf("%s\n", section.Title))
		output.WriteString(strings.Repeat("=", len(section.Title)))
		output.WriteString("\n")

		// Find the longest key for padding
		maxKeyLength := 0
		for _, field := range section.Content {
			if len(field[0]) > maxKeyLength {
				maxKeyLength = len(field[0])
			}
		}

		// Write fields
		for _, field := range section.Content {
			key := field[0]
			value := field[1]
			padding := strings.Repeat(" ", maxKeyLength-len(key))
			output.WriteString(fmt.Sprintf("%s%s: %s\n", key, padding, value))
		}
	}

	return output.String()
}
