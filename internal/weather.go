package internal

import (
	"encoding/json"
	"math/rand"
	"strings"
	"sync"

	"github.com/innotechdevops/openmeteo"
)

type LatLng struct {
	Lat float32
	Lng float32
}

type WeatherOptions struct {
	Location      LatLng
	RelevantHours HoursOptions
	Chart         ChartOptions
}

func (c Config) GetWeatherOptions() WeatherOptions {
	return WeatherOptions{
		Location:      LatLng(c.Weather.Location),
		RelevantHours: c.Weather.Precipitations.Hours.ToHoursOptions(),
		Chart:         c.Weather.Precipitations.Chart.ToChartOptions(),
	}
}

// WeatherInfo holds weather information for the current day.
// The loader func should be used to populate the data.
type WeatherInfo struct {
	Condition      string
	MaxTemperature int
	MinTemperature int

	loader func(*WeatherInfo) error
}

// Load populates the WeatherInfo by calling its loader function.
func (w *WeatherInfo) Load() error {
	if w.loader != nil {
		return w.loader(w)
	}
	return nil
}

// NewWeatherCardAndInfo creates a new weather Card and WeatherInfo using the given latitude and longitude.
func NewWeatherCardAndInfo(options WeatherOptions) (Card, WeatherInfo) {
	var once sync.Once
	var weather weatherData
	var err error

	return makeWeatherCardAndInfo(options, func() (weatherData, error) {
		once.Do(func() {
			weather, err = fetchWeatherData(options.Location)
		})
		return weather, err
	})
}

// NewFakeWeatherCardAndInfo creates a new weather Card and WeatherInfo with fake data for testing purposes.
func NewFakeWeatherCardAndInfo(options WeatherOptions) (Card, WeatherInfo) {
	return makeWeatherCardAndInfo(options, func() (weatherData, error) {
		condition := func() string {
			nth := rand.Int() % len(conditionToIcon)
			i := 0
			for key := range conditionToIcon {
				if i == nth {
					return key
				}
				i++
			}
			return "clear-sky"
		}()

		max := rand.Int()%50 - 20
		min := max - (rand.Int() % 15)
		return weatherData{
			MaxTemperature:                   max,
			MinTemperature:                   min,
			Condition:                        condition,
			HourlyPrecipitationProbabilities: getBiasedSmoothRandomValues(24, 10, 100),
		}, nil
	})
}

func makeWeatherCardAndInfo(options WeatherOptions, getWeather func() (weatherData, error)) (Card, WeatherInfo) {
	return Card{
			Title:    "Précipitations",
			Type:     CardTypeChart,
			Priority: 60,
			loader: func(c *Card) error {
				c.Chart = Chart{}
				data, err := getWeather()
				if err != nil {
					return err
				}

				c.Chart = Chart{
					Data:    data.HourlyPrecipitationProbabilities,
					Hours:   options.RelevantHours,
					Options: options.Chart,
				}
				return nil
			},
		}, WeatherInfo{
			loader: func(w *WeatherInfo) error {
				data, err := getWeather()
				if err != nil {
					return err
				}

				w.Condition = conditionToIcon[data.Condition]
				w.MaxTemperature = data.MaxTemperature
				w.MinTemperature = data.MinTemperature
				return nil
			},
		}
}

type weatherData struct {
	Condition                        string
	MaxTemperature                   int
	MinTemperature                   int
	HourlyPrecipitationProbabilities []int
}

func fetchWeatherData(location LatLng) (weatherData, error) {
	var result weatherData

	param := openmeteo.Parameter{
		Latitude:  openmeteo.Float32(location.Lat),
		Longitude: openmeteo.Float32(location.Lng),
		Timezone:  openmeteo.String("auto"),
		Daily: &[]string{
			openmeteo.DailyWeatherCode,
			openmeteo.DailyTemperature2mMin,
			openmeteo.DailyTemperature2mMax,
		},
		Hourly: &[]string{
			openmeteo.HourlyPrecipitationProbability,
		},
		ForecastDays: openmeteo.Int(1),
	}

	m := openmeteo.New()
	resp, err := m.Execute(param)
	if err != nil {
		return result, err
	}

	var response struct {
		Daily struct {
			Condition []int     `json:"weathercode"`
			MinTemps  []float64 `json:"temperature_2m_min"`
			MaxTemps  []float64 `json:"temperature_2m_max"`
		} `json:"daily"`
		Hourly struct {
			PrecipitationProbs []int `json:"precipitation_probability"`
		} `json:"hourly"`
	}

	err = json.NewDecoder(strings.NewReader(resp)).Decode(&response)
	if err != nil {
		return result, err
	}

	result.Condition = openmeteo.WeatherCodeName(response.Daily.Condition[0])
	result.MaxTemperature = int(response.Daily.MaxTemps[0])
	result.MinTemperature = int(response.Daily.MinTemps[0])
	result.HourlyPrecipitationProbabilities = response.Hourly.PrecipitationProbs

	return result, nil
}

// Map of Open-Meteo weather condition names to the SVG names in the HTML.
var conditionToIcon = map[string]string{
	"clear-sky":                          "sunny",
	"mainly-clear":                       "sunny",
	"partly-cloudy":                      "cloudy",
	"overcast":                           "overcast",
	"fog":                                "foggy",
	"depositing-rime-fog":                "foggy",
	"drizzle-light":                      "rainy",
	"drizzle-moderate":                   "rainy",
	"drizzle-dense":                      "rainy",
	"freezing-drizzle-light":             "rainy",
	"freezing-drizzle-dense":             "rainy",
	"rain-slight":                        "rainy",
	"rain-moderate":                      "rainy",
	"rain-heavy":                         "rainy",
	"freezing-rain-light":                "rainy",
	"freezing-rain-heavy":                "rainy",
	"snow-fall-slight":                   "snowy",
	"snow-fall-moderate":                 "snowy",
	"snow-fall-heavy":                    "snowy",
	"snow-grains":                        "snowy",
	"rain-showers-slight":                "rainy",
	"rain-showers-moderate":              "rainy",
	"rain-showers-violent":               "rainy",
	"snow-showers-slight":                "snowy",
	"snow-showers-heavy":                 "snowy",
	"thunderstorm-slight-or-moderate":    "stormy",
	"thunderstorm-slight-and-heavy-hail": "stormy",
}
