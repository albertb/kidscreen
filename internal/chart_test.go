package internal

import "testing"

func TestChartMaxValue(t *testing.T) {
	if got := (Chart{}).MaxValue(); got != 0 {
		t.Errorf("empty MaxValue() = %d, want 0", got)
	}
	if got := (Chart{Data: []int{42}}).MaxValue(); got != 42 {
		t.Errorf("single-value MaxValue() = %d, want 42", got)
	}
	if got := (Chart{Data: []int{10, 50, 30}}).MaxValue(); got != 50 {
		t.Errorf("multi-value MaxValue() = %d, want 50", got)
	}
}

func TestChartValid(t *testing.T) {
	opts := ChartOptions{Min: 50}
	hours := HoursOptions{Start: 7, End: 20}

	// No data.
	if (Chart{Options: opts, Hours: hours}).Valid() {
		t.Error("empty chart Valid() = true, want false")
	}

	// All hourly values below Min.
	low := make([]int, 24)
	for i := range low {
		low[i] = 30
	}
	if (Chart{Data: low, Options: opts, Hours: hours}).Valid() {
		t.Error("all-below-min chart Valid() = true, want false")
	}

	// One value above Min within the relevant hours.
	above := make([]int, 24)
	above[10] = 80
	if !(Chart{Data: above, Options: opts, Hours: hours}).Valid() {
		t.Error("chart with value 80 at hour 10 Valid() = false, want true")
	}

	// High value only outside the relevant hours.
	outside := make([]int, 24)
	outside[6] = 80 // hour 6 is before Start=7
	if (Chart{Data: outside, Options: opts, Hours: hours}).Valid() {
		t.Error("high value outside range Valid() = true, want false")
	}
}
