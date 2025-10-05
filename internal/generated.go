package internal

import (
	"context"
	"fmt"
	"html/template"
	"math/rand"
	"sync"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// GeneratedOptions holds options for creating LLM-generated Cards.
type GeneratedOptions struct {
	OpenAIAPIKey string
	Cards        []GeneratedCardOptions
}

// GeneratedCardOptions holds options for creating a single LLM-generated Card.
type GeneratedCardOptions struct {
	Title    string
	Prompt   string
	Priority int
}

func (c Config) GetGeneratedOptions() GeneratedOptions {
	return GeneratedOptions{
		OpenAIAPIKey: c.Generated.OpenAIAPIKey,
		Cards: func() []GeneratedCardOptions {
			var cards []GeneratedCardOptions
			for _, card := range c.Generated.Cards {
				cards = append(cards, GeneratedCardOptions(card))
			}
			return cards
		}(),
	}
}

// NewGeneratedCards creates new LLM-generated Cards using the given options.
func NewGeneratedCards(options GeneratedOptions) []Card {
	client := openai.NewClient(
		option.WithAPIKey(options.OpenAIAPIKey),
	)

	var cards []Card
	for _, card := range options.Cards {
		var once sync.Once
		var body string
		var err error

		cards = append(cards, Card{
			Title:    template.HTML(card.Title),
			Type:     CardTypeText,
			Priority: card.Priority,
			loader: func(c *Card) error {
				c.Body = ""
				once.Do(func() {
					body, err = fetchCompletion(client, card.Prompt)
				})
				if err != nil {
					return err
				}
				c.Body = template.HTML(body)
				return nil
			},
		})
	}
	return cards
}

// NewFakeGeneratedCards creates new Cards with hardcoded content for testing purposes.
func NewFakeGeneratedCards() []Card {
	blurbs := []string{
		"En 1900, le premier zeppelin a effectué son vol inaugural. C'était le début de l'ère des dirigeables.",
		"En 1969, l'homme a marché sur la Lune pour la première fois. Un petit pas pour l'homme, un grand pas pour l'humanité.",
		"En 1789, la Révolution française a commencé avec la prise de la Bastille. Liberté, égalité, fraternité!",
		"En 1492, Christophe Colomb a découvert l'Amérique. Un nouveau monde s'est ouvert.",
		"En 1879, Thomas Edison a inventé l'ampoule électrique. La nuit n'a plus jamais été la même.",
	}
	return []Card{
		{
			Title:    "Dans l'histoire",
			Type:     CardTypeText,
			Priority: 60,
			loader: func(c *Card) error {
				c.Body = template.HTML(blurbs[rand.Int()%len(blurbs)])
				return nil
			},
		},
		{
			Title:    "Blague du jour",
			Type:     CardTypeText,
			Priority: 50,
			Body:     "Pet et Répète sont dans un bateau. Pet tombe à l’eau, qui est-ce qui reste?",
		},
	}
}

func fetchCompletion(client openai.Client, prompt string) (string, error) {
	var completion string

	response, err := client.Chat.Completions.New(
		context.Background(),
		openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(fmt.Sprintf("The current date is %s", time.Now().Format("January 2, 2006"))),
				openai.UserMessage(prompt),
			},
			Model: openai.ChatModelGPT4o,
		},
	)
	if err != nil {
		return completion, err
	}
	return response.Choices[0].Message.Content, nil
}
