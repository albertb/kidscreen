package internal

import (
	"testing"
	"time"
)

func TestHeaderFrenchTitle(t *testing.T) {
	tests := []struct {
		date time.Time
		want string
	}{
		// Monday 5 January 2026
		{time.Date(2026, time.January, 5, 12, 0, 0, 0, time.UTC), "Lundi 5 janvier"},
		// Wednesday 4 March 2026
		{time.Date(2026, time.March, 4, 12, 0, 0, 0, time.UTC), "Mercredi 4 mars"},
		// Saturday 14 February 2026
		{time.Date(2026, time.February, 14, 12, 0, 0, 0, time.UTC), "Samedi 14 février"},
	}

	for _, tt := range tests {
		fixed := tt.date
		h := makeHeaderCard(WeatherInfo{}, func() time.Time { return fixed })
		if err := h.Load(); err != nil {
			t.Fatalf("Load() error: %v", err)
		}
		if got := string(h.Title); got != tt.want {
			t.Errorf("date %s: Title = %q, want %q", fixed.Format("2006-01-02"), got, tt.want)
		}
	}
}
