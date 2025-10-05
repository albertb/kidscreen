package internal

import (
	"cmp"
	"os"
	"slices"
)

func Run(config Config, dev, fake bool, img, addr string) error {
	var cards []Card

	cards = append(cards, func() Card {
		options := config.GetAirQualityOptions()
		if fake {
			return NewFakeAirQualityCard(options)
		}
		return NewAirQualityCard(options)
	}())

	err := func() error {
		if fake {
			cards = append(cards, NewFakeCalendarCards()...)
			return nil
		}

		options, err := config.GetCalendarOptions()
		if err != nil {
			return err
		}
		cards = append(cards, NewCalendarCards(options)...)
		return nil
	}()
	if err != nil {
		return err
	}

	cards = append(cards, func() []Card {
		if fake {
			return NewFakeGeneratedCards()
		}
		options := config.GetGeneratedOptions()
		if len(options.Cards) > 0 {
			return NewGeneratedCards(options)
		}
		return nil
	}()...)

	var weather WeatherInfo
	cards = append(cards, func() Card {
		var card Card
		if fake {
			card, weather = NewFakeWeatherCardAndInfo(config.GetWeatherOptions())
			return card
		}
		card, weather = NewWeatherCardAndInfo(config.GetWeatherOptions())
		return card
	}())

	cards = append(cards, func() Card {
		return NewPictureCard(config.GetPictureOptions())
	}())

	header := func() Header {
		if fake {
			return NewFakeHeader(weather)
		}
		return NewHeader(weather)
	}()

	if fake {
		// Add a few filler cards in dev.
		cards = append(cards, []Card{
			{Title: "Disco!", Type: CardTypeText, Body: "J'ai mal au coeur", Priority: 1},
			{Title: "J'appele docteur", Type: CardTypeText, Body: "Il est venu", Priority: 1},
			{Title: "Il est parti", Type: CardTypeText, Body: "En Australie", Priority: 1},
		}...)
	}

	// Sort higher priority cards first.
	slices.SortFunc(cards, func(a, b Card) int {
		return -1 * cmp.Compare(a.Priority, b.Priority)
	})

	if dev {
		DevRender(header, cards, addr)
	} else {
		buf, err := Render(header, cards)
		if err != nil {
			return err
		}
		os.WriteFile(img, buf, 0644)
	}

	return nil
}
