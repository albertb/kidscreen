package internal

import (
	"bytes"
	"testing"
)

// pngMagic is the 8-byte PNG signature.
var pngMagic = []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}

func TestRender(t *testing.T) {
	_, weather := NewFakeWeatherCardAndInfo(WeatherOptions{})
	header := NewFakeHeader(weather)

	var cards []Card
	cards = append(cards, NewFakeAirQualityCard(AirQualityOptions{}))
	cards = append(cards, NewFakeCalendarCards()...)
	cards = append(cards, NewFakeGeneratedCards()...)

	buf, err := Render(header, cards, 1280, 720)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	if !bytes.HasPrefix(buf, pngMagic) {
		t.Errorf("output does not start with PNG magic bytes (got %x)", buf[:min(8, len(buf))])
	}
}
