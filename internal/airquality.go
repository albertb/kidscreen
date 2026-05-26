package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type AirQualityOptions struct {
	Location      LatLng
	RelevantHours HoursOptions
	Chart         ChartOptions
}

func (c Config) GetAirQualityOptions() AirQualityOptions {
	return AirQualityOptions{
		Location:      LatLng(c.Weather.Location),
		RelevantHours: c.Weather.AirQuality.Hours.ToHoursOptions(),
		Chart:         c.Weather.AirQuality.Chart.ToChartOptions(),
	}
}

// NewAirQualityCard creates a new air quality Card using the given latitude and longitude.
func NewAirQualityCard(options AirQualityOptions) Card {
	return makeAirQualityCard(options, newLazy(func() ([]int, error) {
		return fetchAirQualityData(options.Location)
	}))
}

// NewFakeAirQualityCard creates a new air quality Card with fake data for testing purposes.
func NewFakeAirQualityCard(options AirQualityOptions) Card {
	return makeAirQualityCard(options, func() ([]int, error) {
		return getBiasedSmoothRandomValues(24, 0, 250), nil
	})
}

func makeAirQualityCard(options AirQualityOptions, getAQI func() ([]int, error)) Card {
	return Card{
		Title:    "Qualité de l'air",
		Type:     CardTypeChart,
		Priority: 75,
		loader: func(c *Card) error {
			c.Chart = Chart{}
			data, err := getAQI()
			if err != nil {
				return err
			}

			c.Chart = Chart{
				Data:    data,
				Hours:   options.RelevantHours,
				Options: options.Chart,
			}

			return nil
		},
	}
}

func fetchAirQualityData(location LatLng) ([]int, error) {
	var result []int

	url := fmt.Sprintf(
		"https://air-quality-api.open-meteo.com/v1/air-quality?latitude=%f&longitude=%f&hourly=us_aqi&forecast_days=1&timezone=auto",
		location.Lat, location.Lng,
	)

	resp, err := http.Get(url)
	if err != nil {
		return result, fmt.Errorf("failed to get air quality data: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("failed to get air quality data: %s", string(body))
	}

	var response struct {
		Hourly struct {
			AQI []int `json:"us_aqi"`
		} `json:"hourly"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return result, fmt.Errorf("failed to unmarshal air quality JSON: %w", err)
	}
	return response.Hourly.AQI, nil
}
