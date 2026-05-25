package internal

import (
	"errors"
	"fmt"
	"io"
	"time"

	"go.yaml.in/yaml/v4"
)

// Config holds the configuration for the application, parsed from a YAML file.
type Config struct {
	Calendars []Calendar `yaml:"calendars"`
	Weather   Weather    `yaml:"weather"`
	Picture   Picture    `yaml:"picture"`
	Generated Generated  `yaml:"generated"`
	Screen    Screen     `yaml:"screen"`
}

type Calendar struct {
	URL             string `yaml:"url"`
	AttendeesRegExp string `yaml:"attendees_regexp"`
}

type Weather struct {
	Location             Location       `yaml:"location"`
	Precipitations       Precipitations `yaml:"precipitations"`
	AirQuality           AirQuality     `yaml:"airquality"`
	MinDiffThreshold     int            `yaml:"min_diff_threshold"`        // The minimum temperature difference between yesterday and today to display a message about it.
	MinRainfallThreshold int            `yaml:"min_rainfall_threshold_mm"` // The minimum rainfall (mm) to display a message about it.
	MinSnowfallThreshold int            `yaml:"min_snowfall_threshold_cm"` //// The minimum rainfall (cm) to display a message about it.
}

type Location struct {
	Lat float64 `yaml:"lat"`
	Lng float64 `yaml:"lng"`
}

type Precipitations struct {
	Hours TimeRangeConfig `yaml:"relevant_time"`
	Chart ChartConfig     `yaml:"chart"`
}

type AirQuality struct {
	Hours TimeRangeConfig `yaml:"relevant_time"`
	Chart ChartConfig     `yaml:"chart"`
}

type Picture struct {
	PageURL    string `yaml:"page_url"`
	ImageXPath string `yaml:"image_xpath"`
	LabelXPath string `yaml:"label_xpath"`
}

type Generated struct {
	OpenAIAPIKey string          `yaml:"open_ai_api_key"`
	Model        string          `yaml:"model"`
	Cards        []GeneratedCard `yaml:"cards"`
}

type Screen struct {
	Width  int `yaml:"width"`
	Height int `yaml:"height"`
}

type GeneratedCard struct {
	Title    string `yaml:"title"`
	Prompt   string `yaml:"prompt"`
	Priority int    `yaml:"priority"`
}

type TimeRangeConfig struct {
	Start time.Duration `yaml:"start"` // Inclusive
	End   time.Duration `yaml:"end"`   // Inclusive
}

type ChartConfig struct {
	Top  int `yaml:"top"`  // The default top value for the chart if no data exceeds Top.
	Step int `yaml:"step"` // The step size to increase the top value by if some data exceeds it.
	Min  int `yaml:"min"`  // The minimum value for the chart to display.
	High int `yaml:"high"` // The value of maximum shade on the chart.
}

func defaultConfig() Config {
	return Config{
		Screen: Screen{
			Width:  1280,
			Height: 720,
		},
		Weather: Weather{
			Precipitations: Precipitations{
				Hours: TimeRangeConfig{
					Start: 7 * time.Hour,
					End:   20 * time.Hour,
				},
				Chart: ChartConfig{
					Top:  100,
					Step: 5,
					Min:  50,
					High: 75,
				},
			},
			AirQuality: AirQuality{
				Hours: TimeRangeConfig{
					Start: 7 * time.Hour,
					End:   20 * time.Hour,
				},
				Chart: ChartConfig{
					Top:  100,
					Step: 25,
					Min:  45,
					High: 100,
				},
			},
			MinDiffThreshold:     7,
			MinRainfallThreshold: 5,
			MinSnowfallThreshold: 5,
		},
	}
}

func ReadConfig(reader io.Reader) (Config, error) {
	config := defaultConfig()
	if err := yaml.NewDecoder(reader).Decode(&config); err != nil {
		return config, err
	}
	if err := config.validate(); err != nil {
		return config, err
	}
	return config, nil
}

func (c Config) validate() error {
	var errs []error

	for i, cal := range c.Calendars {
		if cal.URL == "" {
			errs = append(errs, fmt.Errorf("calendar[%d]: url is required", i))
		}
	}

	if c.Weather.Location.Lat < -90 || c.Weather.Location.Lat > 90 {
		errs = append(errs, fmt.Errorf("weather.location.lat must be between -90 and 90, got %g", c.Weather.Location.Lat))
	}
	if c.Weather.Location.Lng < -180 || c.Weather.Location.Lng > 180 {
		errs = append(errs, fmt.Errorf("weather.location.lng must be between -180 and 180, got %g", c.Weather.Location.Lng))
	}

	if c.Weather.MinDiffThreshold < 0 {
		errs = append(errs, fmt.Errorf("weather.min_diff_threshold must be non-negative, got %d", c.Weather.MinDiffThreshold))
	}
	if c.Weather.MinRainfallThreshold < 0 {
		errs = append(errs, fmt.Errorf("weather.min_rainfall_threshold_mm must be non-negative, got %d", c.Weather.MinRainfallThreshold))
	}
	if c.Weather.MinSnowfallThreshold < 0 {
		errs = append(errs, fmt.Errorf("weather.min_snowfall_threshold_cm must be non-negative, got %d", c.Weather.MinSnowfallThreshold))
	}

	if c.Screen.Width <= 0 {
		errs = append(errs, fmt.Errorf("screen.width must be positive, got %d", c.Screen.Width))
	}
	if c.Screen.Height <= 0 {
		errs = append(errs, fmt.Errorf("screen.height must be positive, got %d", c.Screen.Height))
	}

	if c.Picture.PageURL != "" {
		if c.Picture.ImageXPath == "" {
			errs = append(errs, fmt.Errorf("picture.image_xpath is required when picture.page_url is set"))
		}
		if c.Picture.LabelXPath == "" {
			errs = append(errs, fmt.Errorf("picture.label_xpath is required when picture.page_url is set"))
		}
	}

	if len(c.Generated.Cards) > 0 && c.Generated.OpenAIAPIKey == "" {
		errs = append(errs, fmt.Errorf("generated.open_ai_api_key is required when generated cards are configured"))
	}

	return errors.Join(errs...)
}
