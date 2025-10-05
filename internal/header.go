package internal

import (
	"html/template"
	"math/rand"
	"strings"
	"time"
)

// Header holds information for the header section of the screen.
// The loader func should be used to populate the data.
type Header struct {
	Title template.HTML

	ConditionSVG   template.HTML
	MaxTemperature int
	MinTemperature int

	loader func(*Header) error
}

// Load populates the Header by calling its loader function.
func (h *Header) Load() error {
	if h.loader != nil {
		return h.loader(h)
	}
	return nil
}

// To translate the date into French.
var replacer = strings.NewReplacer(
	// Days of the week.
	"Sunday", "Dimanche",
	"Monday", "Lundi",
	"Tuesday", "Mardi",
	"Wednesday", "Mercredi",
	"Thursday", "Jeudi",
	"Friday", "Vendredi",
	"Saturday", "Samedi",

	// Months.
	"January", "janvier",
	"February", "février",
	"March", "mars",
	"April", "avril",
	"May", "mai",
	"June", "juin",
	"July", "juillet",
	"August", "août",
	"September", "septembre",
	"October", "octobre",
	"November", "novembre",
	"December", "décembre",
)

// NewHeader creates a new Header using the given WeatherInfo.
func NewHeader(weather WeatherInfo) Header {
	return makeHeaderCard(weather, func() time.Time {
		return time.Now()
	})
}

// NewFakeHeader creates a new Header using the given WeatherInfo with a random date for testing purposes.
func NewFakeHeader(weather WeatherInfo) Header {
	return makeHeaderCard(weather, func() time.Time {
		return time.Now().Add(24 * time.Hour * time.Duration(rand.Intn(364)))
	})
}

func makeHeaderCard(weather WeatherInfo, getTime func() time.Time) Header {
	return Header{
		loader: func(h *Header) error {
			now := getTime()
			if err := weather.Load(); err != nil {
				return err
			}

			h.Title = template.HTML(replacer.Replace(now.Format("Monday 2 January")))
			h.ConditionSVG = template.HTML(weather.Condition)
			h.MaxTemperature = weather.MaxTemperature
			h.MinTemperature = weather.MinTemperature
			return nil
		},
	}
}
