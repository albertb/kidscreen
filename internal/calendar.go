package internal

import (
	"cmp"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"slices"
	"sync"
	"time"

	"github.com/apognu/gocal"
)

// CalendarOptions holds options for creating calendar Cards.
// Attendees can be nil to disable filtering.
type CalendarOptions struct {
	URL       string
	Attendees *regexp.Regexp
}

func (c Config) GetCalendarOptions() ([]CalendarOptions, error) {
	var calendars []CalendarOptions
	for _, cal := range c.Calendars {
		var filter *regexp.Regexp
		if len(cal.AttendeesRegExp) != 0 {
			var err error
			filter, err = regexp.Compile(cal.AttendeesRegExp)
			if err != nil {
				return calendars, fmt.Errorf("failed to compile regex: %w", err)
			}

		}
		calendars = append(calendars, CalendarOptions{URL: cal.URL, Attendees: filter})
	}
	return calendars, nil
}

// MatchesFilter returns true if the event matches the attendees filter, or if no filter is set.
func (c CalendarOptions) MatchesFilter(event gocal.Event) bool {
	if c.Attendees == nil {
		return true
	}
	for _, a := range event.Attendees {
		// Try to match either the name, or contact for each attendee.
		if c.Attendees.MatchString(a.Cn) || c.Attendees.MatchString(a.Value) {
			return true
		}
	}
	return false
}

// NewCalendarCards creates new calendar Cards using the given options.
func NewCalendarCards(options []CalendarOptions) []Card {
	var once sync.Once
	var cal calendar
	var err error

	return makeCalendarCards(func() (calendar, error) {
		once.Do(func() {
			cal, err = fetchCalendars(options)
		})
		return cal, err
	})
}

// NewFakeCalendarCards creates new calendar Cards with fake data for testing purposes.
func NewFakeCalendarCards() []Card {
	var (
		mu     sync.Mutex
		cached calendar
		last   time.Time
	)
	return makeCalendarCards(func() (calendar, error) {
		mu.Lock()
		defer mu.Unlock()

		if time.Since(last) < time.Second {
			return cached, nil
		}

		now := time.Now()
		midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		today := []event{}
		tomorrow := []event{}

		for _, e := range []event{
			{Summary: "Estelle: Arts plastiques"},
			{Summary: "Julie: Musique"},
			{Summary: "Parc avec les amis", Time: midnight.Add(10*time.Hour + 30*time.Minute)},
			{Summary: "Dîner au restaurant", Time: midnight.Add(12 * time.Hour)},
			{Summary: "Souper chez mamie", Time: midnight.Add(18 * time.Hour)},
			{Summary: "Film en famille", Time: midnight.Add(19 * time.Hour)},
		} {
			x := rand.Float32()
			switch {
			case x > 0.75:
				today = append(today, e)
			case x > 0.5 && x <= 0.75:
				tomorrow = append(tomorrow, e)
			}
		}

		cached = calendar{
			Today:    today,
			Tomorrow: tomorrow,
		}
		last = time.Now()
		return cached, nil
	})
}

type calendar struct {
	Today    []event
	Tomorrow []event
}

type event struct {
	Summary string
	Time    time.Time
}

func fetchCalendars(options []CalendarOptions) (calendar, error) {
	var calendar calendar

	for _, c := range options {
		resp, err := http.Get(c.URL)
		if err != nil {
			return calendar, err
		}
		defer resp.Body.Close()

		now := time.Now()

		// We only care about events today and tomorrow.
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		tomorrow := today.Add(24 * time.Hour)

		// To catch all-day events, let them start one second before today.
		start := today.Add(-1 * time.Second)
		end := tomorrow.Add(24 * time.Hour)

		// Parse events for today and tomorrow.
		cal := gocal.NewParser(resp.Body)
		cal.AllDayEventsTZ = now.Location()
		cal.Start, cal.End = &start, &end

		err = cal.Parse()
		if err != nil {
			return calendar, err
		}

		for _, e := range cal.Events {
			// Ignore all-day events from the previous day.
			if e.Start.Before(today) {
				continue
			}

			fmt.Println("Event:", e.Summary, "Start:", e.Start)

			// Only consider events that match the attendees filter.
			if !c.MatchesFilter(e) {
				fmt.Println("  (skipping, doesn't match filter)")
				continue
			}

			// Don't include a time for all day events.
			var time time.Time
			if e.Start.Hour() > 0 || e.Start.Minute() > 0 {
				time = *e.Start
			}

			if e.Start.Before(tomorrow) {
				calendar.Today = append(calendar.Today, event{
					Summary: e.Summary,
					Time:    time,
				})
			} else {
				calendar.Tomorrow = append(calendar.Tomorrow, event{
					Summary: e.Summary,
					Time:    time,
				})
			}
		}
	}

	// Sort the events from all the calendars.
	byStart := func(a, b event) int {
		return cmp.Compare(a.Time.Unix(), b.Time.Unix())
	}
	slices.SortFunc(calendar.Today, byStart)
	slices.SortFunc(calendar.Tomorrow, byStart)

	fmt.Println("Today: ", calendar.Today)
	fmt.Println("Tomorrow: ", calendar.Tomorrow)

	return calendar, nil
}

func (e event) String() string {
	time := ""
	// All-day events don't have a time, so will just show the summary.
	if !e.Time.IsZero() {
		time = e.Time.Local().Format("15h04 ")
	}
	return time + e.Summary
}

func makeCalendarCards(getCalendar func() (calendar, error)) []Card {
	return []Card{
		{
			Title:    "Aujourd'hui",
			Priority: 100,
			loader: func(c *Card) error {
				calendar, err := getCalendar()
				if err != nil {
					return err
				}

				c.Body = ""
				c.Items = []string{}

				if len(calendar.Today) == 0 {
					return nil
				}

				c.Type = CardTypeList
				for _, e := range calendar.Today {
					c.Items = append(c.Items, e.String())
				}

				return nil
			},
		},
		{
			Title:    "Demain",
			Type:     CardTypeList,
			Priority: 50,
			loader: func(c *Card) error {
				calendar, err := getCalendar()
				if err != nil {
					return err
				}

				c.Items = []string{}
				if len(calendar.Tomorrow) == 0 {
					return nil
				}
				for _, e := range calendar.Tomorrow {
					c.Items = append(c.Items, e.String())
				}

				return nil
			},
		},
	}
}
