package internal

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"

	"github.com/chromedp/chromedp"
)

//go:embed screen.go.html
var templates embed.FS

type renderData struct {
	Header *Header
	Cards  []*Card
}

func assembleData(header Header, cards []Card) renderData {
	var data renderData

	// Load the header and cards in parallel since most involve network calls.
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := header.Load(); err != nil {
			log.Printf("failed to load header: %v", err)
		}
	}()

	for i := range cards {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := cards[i].Load(); err != nil {
				log.Printf("failed to load card %q: %v", cards[i].Title, err)
			}
		}()
	}

	wg.Wait()

	data.Header = &header
	for _, card := range cards {
		if card.Valid() {
			data.Cards = append(data.Cards, &card)
		}
	}
	return data
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
		data := assembleData(header, cards)
		if err := tmpl.Execute(w, data); err != nil {
			log.Println("failed to execute template:", err)
		}
	})
	go func() {
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Println("server error:", err)
		}
	}()
	fmt.Printf("Server running on http://localhost%s/\nPress enter to stop\n", addr)
	fmt.Scanln()
}

// Render generates a PNG image of the screen with the given header and cards.
func Render(header Header, cards []Card, width, height int) ([]byte, error) {
	tmpl, err := template.ParseFS(templates, "screen.go.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	errc := make(chan error, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := assembleData(header, cards)
		if err := tmpl.Execute(w, data); err != nil {
			select {
			case errc <- fmt.Errorf("failed to execute template: %w", err):
			default:
			}
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
		chromedp.EmulateViewport(int64(width), int64(height)),
		chromedp.Navigate(ts.URL),
		chromedp.WaitReady("main"),
		chromedp.OuterHTML("html", &body),
		chromedp.CaptureScreenshot(&buf),
	); err != nil {
		return nil, fmt.Errorf("failed to capture screenshot: %w", err)
	}

	select {
	case err := <-errc:
		return nil, err
	default:
		return buf, nil
	}
}
