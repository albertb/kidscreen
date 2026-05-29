package internal

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"net/url"
	"time"

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
	if options.PageURL == "" {
		return Card{}
	}

	fetch := newLazy(withRetry(3, func() (picture, error) {
		return fetchPicture(options)
	}))

	return Card{
		Type:     CardTypeText,
		Priority: 35,
		loader: func(c *Card) error {
			pic, err := fetch()
			if err != nil {
				return err
			}
			c.Title = template.HTML(pic.Label)
			c.Body = template.HTML(fmt.Sprintf(`<img src="%s">`, pic.URL))
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

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(options.PageURL)
	if err != nil {
		return pic, fmt.Errorf("failed to load animal page: %w", err)
	}
	defer resp.Body.Close()

	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		return pic, fmt.Errorf("failed to parse animal page: %w", err)
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

	if len(labels) == 0 {
		return pic, fmt.Errorf("no pictures found on page")
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
