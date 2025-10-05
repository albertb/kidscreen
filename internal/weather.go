package internal

import (
	"encoding/json"
	"fmt"
	"html/template"
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
	Location         LatLng
	MinDiffThreshold int
	RelevantHours    HoursOptions
	Chart            ChartOptions
}

func (c Config) GetWeatherOptions() WeatherOptions {
	return WeatherOptions{
		Location:         LatLng(c.Weather.Location),
		MinDiffThreshold: c.Weather.MinDiffThreshold,
		RelevantHours:    c.Weather.Precipitations.Hours.ToHoursOptions(),
		Chart:            c.Weather.Precipitations.Chart.ToChartOptions(),
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
func NewWeatherCardAndInfo(options WeatherOptions) ([]Card, WeatherInfo) {
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
func NewFakeWeatherCardAndInfo(options WeatherOptions) ([]Card, WeatherInfo) {
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

		maxToday := rand.Int()%50 - 20
		minToday := maxToday - (rand.Int() % 15)

		maxYesterday := maxToday + 15 - (rand.Int() % 30)
		minYesterday := maxYesterday - (rand.Int() % 15)

		return weatherData{
			TemperatureToday: temperatureData{
				Max: maxToday,
				Min: minToday,
			},
			TemperatureYesterday: temperatureData{
				Max: maxYesterday,
				Min: minYesterday,
			},
			Condition:                        condition,
			HourlyPrecipitationProbabilities: getBiasedSmoothRandomValues(24, 10, 100),
		}, nil
	})
}

func makeWeatherCardAndInfo(options WeatherOptions, getWeather func() (weatherData, error)) ([]Card, WeatherInfo) {
	return []Card{
			{
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
			},
			{
				Title:    "Température",
				Type:     CardTypeText,
				Priority: 60,
				loader: func(c *Card) error {
					data, err := getWeather()
					if err != nil {
						return err
					}

					diff := data.TemperatureToday.Max - data.TemperatureYesterday.Max

					if diff > options.MinDiffThreshold {
						c.Body = template.HTML(fmt.Sprintf("%d°C plus chaud qu'hier.", diff))
					} else if diff < -options.MinDiffThreshold {
						c.Body = template.HTML(fmt.Sprintf("%d°C plus froid qu'hier.", -diff))
					}
					return nil
				},
			},
		},
		WeatherInfo{
			loader: func(w *WeatherInfo) error {
				data, err := getWeather()
				if err != nil {
					return err
				}

				w.Condition = conditionToIcon[data.Condition]
				w.MaxTemperature = data.TemperatureToday.Max
				w.MinTemperature = data.TemperatureToday.Min
				return nil
			},
		}
}

type weatherData struct {
	Condition                        string
	TemperatureToday                 temperatureData
	TemperatureYesterday             temperatureData
	HourlyPrecipitationProbabilities []int
}

type temperatureData struct {
	Max int
	Min int
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
		PastDays:     openmeteo.Int(1),
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
	result.TemperatureYesterday.Max = int(response.Daily.MaxTemps[0])
	result.TemperatureYesterday.Min = int(response.Daily.MinTemps[0])
	result.TemperatureToday.Max = int(response.Daily.MaxTemps[1])
	result.TemperatureToday.Min = int(response.Daily.MinTemps[1])
	result.HourlyPrecipitationProbabilities = response.Hourly.PrecipitationProbs[24:]

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
