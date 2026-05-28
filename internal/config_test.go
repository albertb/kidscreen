package internal

import (
	"strings"
	"testing"
)

func TestReadConfigDefaults(t *testing.T) {
	cfg, err := ReadConfig(strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Screen.Width != 1280 || cfg.Screen.Height != 720 {
		t.Errorf("default screen = %dx%d, want 1280x720", cfg.Screen.Width, cfg.Screen.Height)
	}
}

func TestReadConfigInvalidYAML(t *testing.T) {
	_, err := ReadConfig(strings.NewReader(": bad yaml :::"))
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestReadConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name:    "empty calendar URL",
			yaml:    "calendars:\n  - url: \"\"\n",
			wantErr: "calendar[0]: url is required",
		},
		{
			name:    "invalid calendar regexp",
			yaml:    "calendars:\n  - url: http://x\n    attendees_regexp: \"[\"\n",
			wantErr: "invalid attendees_regexp",
		},
		{
			name:    "lat out of range",
			yaml:    "weather:\n  location:\n    lat: 91\n    lng: 0\n",
			wantErr: "weather.location.lat",
		},
		{
			name:    "lng out of range",
			yaml:    "weather:\n  location:\n    lat: 0\n    lng: 181\n",
			wantErr: "weather.location.lng",
		},
		{
			name:    "negative min_diff_threshold",
			yaml:    "weather:\n  min_diff_threshold: -1\n",
			wantErr: "weather.min_diff_threshold",
		},
		{
			name:    "zero screen width",
			yaml:    "screen:\n  width: 0\n  height: 720\n",
			wantErr: "screen.width",
		},
		{
			name:    "picture page_url without image_xpath",
			yaml:    "picture:\n  page_url: http://x\n  label_xpath: //x\n",
			wantErr: "picture.image_xpath",
		},
		{
			name:    "generated cards without API key",
			yaml:    "generated:\n  cards:\n    - title: t\n      prompt: p\n",
			wantErr: "generated.open_ai_api_key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ReadConfig(strings.NewReader(tt.yaml))
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestReadConfigAttendeesRegExp(t *testing.T) {
	yaml := "calendars:\n  - url: http://x\n    attendees_regexp: \"Alice|Bob\"\n"
	cfg, err := ReadConfig(strings.NewReader(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cal := cfg.Calendars[0]
	if cal.AttendeesRegExp == nil {
		t.Fatal("AttendeesRegExp is nil, want compiled regexp")
	}
	if !cal.AttendeesRegExp.MatchString("Alice") {
		t.Error("compiled regexp does not match \"Alice\"")
	}
}
