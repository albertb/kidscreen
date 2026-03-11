package internal

import (
	"context"
	"fmt"
	"html/template"
	"math"
	"math/rand"
	"strings"
	"sync"

	//"github.com/innotechdevops/openmeteo"
	"github.com/hectormalot/omgo"
)

type LatLng struct {
	Lat float64
	Lng float64
}

type WeatherOptions struct {
	Location             LatLng
	MinDiffThreshold     int
	MinRainfallThreshold int // mm
	MinSnowfallThreshold int // cm
	RelevantHours        HoursOptions
	Chart                ChartOptions
}

func (c Config) GetWeatherOptions() WeatherOptions {
	return WeatherOptions{
		Location:             LatLng(c.Weather.Location),
		MinDiffThreshold:     c.Weather.MinDiffThreshold,
		MinRainfallThreshold: c.Weather.MinRainfallThreshold,
		MinSnowfallThreshold: c.Weather.MinSnowfallThreshold,
		RelevantHours:        c.Weather.Precipitations.Hours.ToHoursOptions(),
		Chart:                c.Weather.Precipitations.Chart.ToChartOptions(),
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
					return conditionToIcon[key]
				}
				i++
			}
			return "sunny"
		}()

		maxToday := rand.Int()%50 - 20
		minToday := maxToday - (rand.Int() % 15)

		maxYesterday := maxToday + 15 - (rand.Int() % 30)
		minYesterday := maxYesterday - (rand.Int() % 15)

		rainfall := math.Abs(rand.NormFloat64())*2.0 + 2.0
		snowfall := math.Abs(rand.NormFloat64())*2.0 + 2.0

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
			HourlyPrecipitationProbabilities: getBiasedSmoothRandomValues(24, 10., 100.),
			Rainfall:                         rainfall,
			Snowfall:                         snowfall,
		}, nil
	})
}

func makeWeatherCardAndInfo(options WeatherOptions, getWeather func() (weatherData, error)) ([]Card, WeatherInfo) {
	return []Card{
			{
				Title:    "Précipitations",
				Type:     CardTypeChart,
				Priority: 65,
				loader: func(c *Card) error {
					c.Chart = Chart{}
					data, err := getWeather()
					if err != nil {
						return err
					}

					var probs []int
					for _, prob := range data.HourlyPrecipitationProbabilities {
						probs = append(probs, int(prob))
					}

					c.Chart = Chart{
						Data:    probs,
						Hours:   options.RelevantHours,
						Options: options.Chart,
					}
					return nil
				},
			},
			{
				Title:    "Météo",
				Type:     CardTypeText,
				Priority: 60,
				loader: func(c *Card) error {
					data, err := getWeather()
					if err != nil {
						return err
					}
					sb := strings.Builder{}

					if data.Rainfall > float64(options.MinRainfallThreshold) {
						sb.WriteString(fmt.Sprintf("%3.1f mm de pluie<br>", data.Rainfall))
					}
					if data.Snowfall > float64(options.MinSnowfallThreshold) {
						sb.WriteString(fmt.Sprintf("%3.1f cm de neige<br>", data.Snowfall))
					}

					diff := data.TemperatureToday.Max - data.TemperatureYesterday.Max
					if diff > options.MinDiffThreshold {
						sb.WriteString(fmt.Sprintf("%d°C plus chaud qu'hier", diff))
					} else if diff < -options.MinDiffThreshold {
						sb.WriteString(fmt.Sprintf("%d°C plus froid qu'hier", -diff))
					}

					c.Body = template.HTML(sb.String())
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

				w.Condition = data.Condition
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
	HourlyPrecipitationProbabilities []float64
	Rainfall                         float64
	Snowfall                         float64
}

type temperatureData struct {
	Max int
	Min int
}

func fetchWeatherData(location LatLng) (weatherData, error) {
	var result weatherData

	client := omgo.NewClient()

	req, err := omgo.NewForecastRequest(location.Lat, location.Lng)
	if err != nil {
		return result, err
	}

	req.
		WithDaily(
			omgo.DailyWeatherCode,
			omgo.DailyTemperature2mMin,
			omgo.DailyTemperature2mMax,
			omgo.DailyShowersSum,
			omgo.DailySnowfallSum).
		WithHourly(
			omgo.HourlyPrecipitationProbability).
		WithForecastDays(1).
		WithPastDays(1).
		WithTimezone("auto")

	weather, err := client.Forecast(context.Background(), req)
	if err != nil {
		return result, err
	}

	result.Condition = conditionToIcon[weather.Daily.WeatherCode[1]]
	result.TemperatureYesterday.Max = int(weather.Daily.Temperature2mMax[0])
	result.TemperatureYesterday.Min = int(weather.Daily.Temperature2mMin[0])
	result.TemperatureToday.Max = int(weather.Daily.Temperature2mMax[1])
	result.TemperatureToday.Min = int(weather.Daily.Temperature2mMin[1])
	result.HourlyPrecipitationProbabilities = weather.Hourly.PrecipitationProbability[24:]
	result.Rainfall = weather.Daily.ShowersSum[1]
	result.Snowfall = weather.Daily.SnowfallSum[1]

	return result, nil
}

// Map of open-meteo weather condition codes to the SVG names in the HTML.
var conditionToIcon = map[omgo.WeatherCode]string{
	omgo.ClearSky:                   "sunny",
	omgo.MainlyClear:                "sunny",
	omgo.PartlyCloudy:               "cloudy",
	omgo.Overcast:                   "overcast",
	omgo.Fog:                        "foggy",
	omgo.DepositingRimeFog:          "foggy",
	omgo.DrizzleLight:               "rainy",
	omgo.DrizzleModerate:            "rainy",
	omgo.DrizzleDense:               "rainy",
	omgo.FreezingDrizzleLight:       "rainy",
	omgo.FreezingDrizzleDense:       "rainy",
	omgo.RainSlight:                 "rainy",
	omgo.RainModerate:               "rainy",
	omgo.RainHeavy:                  "rainy",
	omgo.FreezingRainLight:          "rainy",
	omgo.FreezingRainHeavy:          "rainy",
	omgo.SnowFallSlight:             "snowy",
	omgo.SnowFallModerate:           "snowy",
	omgo.SnowFallHeavy:              "snowy",
	omgo.SnowGrains:                 "snowy",
	omgo.RainShowersSlight:          "rainy",
	omgo.RainShowersModerate:        "rainy",
	omgo.RainShowersViolent:         "rainy",
	omgo.SnowShowersSlight:          "snowy",
	omgo.SnowShowersHeavy:           "snowy",
	omgo.ThunderstormSlight:         "stormy",
	omgo.ThunderstormWithHailSlight: "stormy",
	omgo.ThunderstormWithHailHeavy:  "stormy",
}
