package internal

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

//go:embed screen.go.html
var templates embed.FS

type renderData struct {
	Header *Header
	Cards  []*Card
}

func assembleData(header Header, cards []Card) (renderData, error) {
	var data renderData

	// Load the header and cards in parallel since most involve network calls.
	var wg sync.WaitGroup
	errs := make(chan error, 1+len(cards))

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := header.Load(); err != nil {
			errs <- fmt.Errorf("failed to load header: %w", err)
		}
	}()

	for i := range cards {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := cards[i].Load(); err != nil {
				errs <- fmt.Errorf("failed to load card (%s): %w", cards[i].Title, err)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(errs)
	}()

	var err error
	for e := range errs {
		err = errors.Join(err, e)
	}
	if err != nil {
		log.Println("failed to load some cards:", err)
	}

	data.Header = &header
	for _, card := range cards {
		if card.Valid() {
			data.Cards = append(data.Cards, &card)
		}
	}
	return data, nil
}

// DevRender starts a local HTTP server to render the screen for development purposes.
func DevRender(header Header, cards []Card, addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("internal/screen.go.html")
		if err != nil {
			log.Println("failed to load template:", err)
			return
		}
		data, err := assembleData(header, cards)
		if err != nil {
			log.Println("failed to load data:", err)
		}

		if err := tmpl.Execute(w, data); err != nil {
			log.Println("failed to execute template:", err)
		}
	})
	go func() { http.ListenAndServe(addr, mux) }()
	fmt.Printf("Server running on http://localhost%s/\nPress enter to stop\n", addr)
	fmt.Scanln()
}

// Render generates a PNG image of the screen with the given header and cards.
func Render(header Header, cards []Card) ([]byte, error) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFS(templates, "screen.go.html")
		if err != nil {
			log.Println(err)
		}

		data, err := assembleData(header, cards)
		if err != nil {
			log.Println("failed to load data:", err)
		}

		if err := tmpl.Execute(w, data); err != nil {
			log.Println(err)
		}
	}))
	defer ts.Close()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoSandbox,
		chromedp.Headless,
		chromedp.DisableGPU,
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var body string
	var buf []byte
	if err := chromedp.Run(ctx,
		chromedp.EmulateViewport(1280, 720),
		chromedp.Navigate(ts.URL),
		chromedp.Sleep(1*time.Second),
		chromedp.OuterHTML("html", &body),
		chromedp.CaptureScreenshot(&buf),
	); err != nil {
		fmt.Println("Press enter to stop")
		fmt.Scanln()

		return []byte{}, fmt.Errorf("failed to capture screenshot: %w", err)
	}

	return buf, nil
}
