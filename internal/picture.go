package internal

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/url"
	"sync"

	"github.com/antchfx/htmlquery"
)

// PictureOptions holds options for creating the picture Card.
type PictureOptions struct {
	PageURL     string // The URL of the page to scrape for pictures.
	ImagesXPath string // The XPath to select the image elements.
	LabelXPath  string // The XPath to select the label element relative to the image element.
}

func (c Config) GetPictureOptions() PictureOptions {
	return PictureOptions{
		PageURL:     c.Picture.PageURL,
		ImagesXPath: c.Picture.ImageXPath,
		LabelXPath:  c.Picture.LabelXPath,
	}
}

// NewPictureCard creates a new picture Card using the given options.
// The card will display a random picture and its label from the specified page.
func NewPictureCard(options PictureOptions) Card {
	var once sync.Once
	var picture picture
	var err error

	if options.PageURL == "" {
		return Card{}
	}

	return Card{
		Type:     CardTypeText,
		Priority: 35,
		loader: func(c *Card) error {
			once.Do(func() {
				picture, err = fetchPicture(options)
			})
			if err != nil {
				return err
			}
			c.Title = template.HTML(picture.Label)
			c.Body = template.HTML(fmt.Sprintf(`<img src="%s">`, picture.URL))
			return nil
		},
	}
}

type picture struct {
	Label string
	URL   string
}

func fetchPicture(options PictureOptions) (picture, error) {
	var pic picture

	doc, err := htmlquery.LoadURL(options.PageURL)
	if err != nil {
		return pic, fmt.Errorf("failed to load animal page: %w", err)
	}

	images, err := htmlquery.QueryAll(doc, options.ImagesXPath)
	if err != nil {
		return pic, fmt.Errorf("failed to query document: %w", err)
	}

	labels := map[string]string{}
	for _, image := range images {
		label := htmlquery.FindOne(image, options.LabelXPath)
		if label == nil {
			continue
		}
		labels[htmlquery.InnerText(label)] = htmlquery.SelectAttr(image, "src")
	}

	i := 0
	nth := rand.Int() % len(labels)
	for label, image := range labels {
		if i >= nth {
			url, err := url.Parse(image)
			if err != nil {
				return pic, fmt.Errorf("failed to parse picture URL: %w", err)
			}
			if !url.IsAbs() {
				page, err := url.Parse(options.PageURL)
				if err != nil {
					return pic, fmt.Errorf("failed to parse picture page URL: %w", err)
				}
				url = page.ResolveReference(url)
			}
			return picture{Label: label, URL: url.String()}, nil
		}
		i++
	}
	return pic, fmt.Errorf("failed to pick a picture")
}
