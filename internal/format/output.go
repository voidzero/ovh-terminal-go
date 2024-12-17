// internal/format/output.go
package format

import (
	"fmt"
	"strings"
)

// Alignment represents text alignment options
type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

// SectionConfig defines configuration options for sections
type SectionConfig struct {
	TitleAlignment  Alignment
	TitleDecorator  string
	KeyAlignment    Alignment
	ValueAlignment  Alignment
	Indent          int
	KeyValueSpacing int
}

// DefaultConfig provides standard formatting configuration
var DefaultConfig = SectionConfig{
	TitleAlignment:  AlignLeft,
	TitleDecorator:  "=",
	KeyAlignment:    AlignLeft,
	ValueAlignment:  AlignLeft,
	Indent:          0,
	KeyValueSpacing: 1,
}

// Field represents a single key-value field
type Field struct {
	Key          string
	Value        string
	ValueLines   []string
	SkipIfEmpty  bool
	IsDecorative bool
}

// Section represents a group of related data
type Section struct {
	Title    string
	Content  []Field
	Config   SectionConfig
	parent   *OutputFormatter
	maxWidth int
}

// OutputFormatter handles formatted text output
type OutputFormatter struct {
	sections  []*Section
	maxWidth  int
	separator string
}

// FormatterOption defines options for the formatter
type FormatterOption func(*OutputFormatter)

// WithMaxWidth sets maximum output width
func WithMaxWidth(width int) FormatterOption {
	return func(f *OutputFormatter) {
		f.maxWidth = width
	}
}

// WithSeparator sets the separator between sections
func WithSeparator(sep string) FormatterOption {
	return func(f *OutputFormatter) {
		f.separator = sep
	}
}

// NewOutputFormatter creates a new formatter instance
func NewOutputFormatter(opts ...FormatterOption) *OutputFormatter {
	f := &OutputFormatter{
		sections:  make([]*Section, 0),
		maxWidth:  80,
		separator: "\n",
	}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

// AddSection adds a new section with optional configuration
func (f *OutputFormatter) AddSection(title string, config ...SectionConfig) *Section {
	cfg := DefaultConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	section := &Section{
		Title:    title,
		Content:  make([]Field, 0),
		Config:   cfg,
		parent:   f,
		maxWidth: f.maxWidth,
	}
	f.sections = append(f.sections, section)
	return section
}

// AddField adds a field to a section
func (s *Section) AddField(key, value string) *Section {
	if value != "" {
		s.Content = append(s.Content, Field{
			Key:         key,
			Value:       value,
			SkipIfEmpty: true,
		})
	}
	return s
}

// AddFields adds multiple fields at once
func (s *Section) AddFields(fields map[string]string) *Section {
	for key, value := range fields {
		s.AddField(key, value)
	}
	return s
}

// AddMultiLineField adds a field with multiple value lines
func (s *Section) AddMultiLineField(key string, values []string) *Section {
	if len(values) > 0 {
		s.Content = append(s.Content, Field{
			Key:         key,
			ValueLines:  values,
			SkipIfEmpty: true,
		})
	}
	return s
}

// AddDivider adds a decorative line
func (s *Section) AddDivider(char string) *Section {
	s.Content = append(s.Content, Field{
		Value:        strings.Repeat(char, s.maxWidth),
		IsDecorative: true,
	})
	return s
}

// align handles text alignment within a given width
func align(text string, alignment Alignment, width int) string {
	textLen := len(text)
	if textLen >= width {
		return text
	}

	spaces := width - textLen
	if spaces <= 0 {
		return text
	}

	switch alignment {
	case AlignCenter:
		leftPad := spaces / 2
		rightPad := spaces - leftPad
		return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
	case AlignRight:
		return strings.Repeat(" ", spaces) + text
	default: // AlignLeft
		return text + strings.Repeat(" ", spaces)
	}
}

// String formats the entire output
func (f *OutputFormatter) String() string {
	var output strings.Builder

	for i, section := range f.sections {
		if i > 0 {
			output.WriteString(f.separator)
		}

		// Write section title
		if section.Title != "" {
			title := align(section.Title, section.Config.TitleAlignment, f.maxWidth)
			output.WriteString(fmt.Sprintf("%s\n", title))
			if section.Config.TitleDecorator != "" {
				output.WriteString(
					strings.Repeat(section.Config.TitleDecorator, len(section.Title)),
				)
				output.WriteString("\n")
			}
		}

		// Find maximum key length for alignment
		maxKeyLength := 0
		for _, field := range section.Content {
			if !field.IsDecorative && len(field.Key) > maxKeyLength {
				maxKeyLength = len(field.Key)
			}
		}

		// Write fields
		indent := strings.Repeat(" ", section.Config.Indent)
		for i, field := range section.Content {
			if field.SkipIfEmpty && field.Value == "" && len(field.ValueLines) == 0 {
				continue
			}

			if field.IsDecorative {
				output.WriteString(field.Value)
			} else {
				key := align(field.Key, section.Config.KeyAlignment, maxKeyLength)
				spacing := strings.Repeat(" ", section.Config.KeyValueSpacing)

				if len(field.ValueLines) > 0 {
					// Handle multi-line values
					output.WriteString(fmt.Sprintf("%s%s%s%s\n", indent, key, spacing, field.ValueLines[0]))
					for _, line := range field.ValueLines[1:] {
						padding := strings.Repeat(" ", maxKeyLength+section.Config.KeyValueSpacing)
						output.WriteString(fmt.Sprintf("%s%s%s\n", indent, padding, line))
					}
				} else {
					// Handle single-line value
					value := align(field.Value, section.Config.ValueAlignment,
						f.maxWidth-maxKeyLength-section.Config.KeyValueSpacing-section.Config.Indent)
					output.WriteString(fmt.Sprintf("%s%s%s%s", indent, key, spacing, value))
				}
			}

			if i < len(section.Content)-1 {
				output.WriteString("\n")
			}
		}
	}

	return output.String()
}
