package internal

import (
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
}

type Calendar struct {
	URL             string `yaml:"url"`
	AttendeesRegExp string `yaml:"attendees_regexp"`
}

type Weather struct {
	Location       Location       `yaml:"location"`
	Precipitations Precipitations `yaml:"precipitations"`
	AirQuality     AirQuality     `yaml:"airquality"`
}

type Location struct {
	Lat float32 `yaml:"lat"`
	Lng float32 `yaml:"lng"`
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
	Cards        []GeneratedCard `yaml:"cards"`
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
	// TODO
	return nil
}
