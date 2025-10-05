package internal

import (
	"html/template"
)

// CardType represents the type of content a Card holds.
type CardType int

const (
	CardTypeUnknown CardType = iota
	CardTypeText             // Supports title, body, and footer.
	CardTypeList             // Supports title, list, and footer.
	CardTypeChart            // Supports title, chart, and footer.
)

// Card represents a single information card to be displayed on the screen.
type Card struct {
	Title  template.HTML
	Footer template.HTML
	Type   CardType

	// For CardTypeText
	Body template.HTML

	// For CardTypeList
	Items []string

	// For CardTypeChart
	Chart Chart

	Priority int

	loader func(*Card) error
}

// Load invokes the loader function to populate the Card's dynamic content.
func (c *Card) Load() error {
	if c.loader != nil {
		return c.loader(c)
	}
	return nil
}

// Returns whether a card is valid and should be displayed.
func (c Card) Valid() bool {
	if c.Type == CardTypeText {
		return len(c.Body) > 0
	}
	if c.Type == CardTypeList {
		return len(c.Items) > 0
	}
	if c.Type == CardTypeChart {
		return c.Chart.Valid()
	}
	return false
}
