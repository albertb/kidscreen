package internal

import (
	"slices"
)

// Chart represents a simple bar chart to be displayed on a Card.
type Chart struct {
	Data    []int        // The points to graph on the chart.
	Hours   HoursOptions // The relevant hours to display; high data outside this range is ignored.
	Options ChartOptions // Chart display options.
}

type HoursOptions struct {
	Start int // Inclusive
	End   int // Inclusive
}

func (c Chart) MaxValue() int {
	return slices.Max(c.Data)
}

func (c Chart) Valid() bool {
	if len(c.Data) == 0 {
		return false
	}
	for hour, value := range c.Data {
		if hour < c.Hours.Start || hour > c.Hours.End {
			continue
		}
		if value > c.Options.Min {
			return true
		}
	}
	return false
}

func (c TimeRangeConfig) ToHoursOptions() HoursOptions {
	return HoursOptions{
		Start: int(c.Start.Hours()),
		End:   int(c.End.Hours()),
	}
}

type ChartOptions struct {
	Top  int // The default top value for the chart if no data exceeds HighValue.
	Step int // The step for the top value.

	Min  int // The minimum value for the chart to display.
	High int // The value of maximum shade.
}

func (c ChartConfig) ToChartOptions() ChartOptions {
	return ChartOptions(c)
}
