package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// FetchersMap stores Fetchers references under keys of the same values
// as their corresponding Fetcher.ID
type FetchersMap map[uint]*Fetcher

// Fetchers wraps FetchersMap with mutex and keeps next ID to be used
type Fetchers struct {
	sync.Mutex
	fetchersMap FetchersMap
	nextIndex   uint
}

// NewFetchers initializes Fetchers with default values
func NewFetchers() *Fetchers {
	fetchers := &Fetchers{
		Mutex:       sync.Mutex{},
		fetchersMap: make(FetchersMap),
		nextIndex:   1,
	}
	return fetchers
}

var fetchers *Fetchers

func init() {
	// add mocks of few initial Fetchers or use empty Fetchers
	// mocks:
	fetchers = &Fetchers{
		fetchersMap: FetchersMap{
			1: &Fetcher{ID: 1, URL: "https://httpbin.org/range/15", Interval: 7},
			2: &Fetcher{ID: 2, URL: "https://httpbin.org/delay/10", Interval: 20},
			3: &Fetcher{ID: 3, URL: "https://httpbin.org/base64/U8WCYXdvbWlyLVphd2lzdG93c2tp", Interval: 10},
		},
		nextIndex: 4,
	}

	for _, f := range fetchers.fetchersMap {
		f.initWorker()
	}

	// empty:
	// fetchers = NewFetchers()
}

func main() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Route("/api/fetcher", func(r chi.Router) {
		r.Get("/", getFetchers)
		r.Post("/", addFetcher)   // add
		r.Put("/", updateFetcher) // update

		r.Route("/{id:\\d+}", func(r chi.Router) {
			r.Get("/", getFetcher)
			r.Delete("/", removeFetcher) // remove

			r.Route("/history", func(r chi.Router) {
				r.Get("/", getFetcherHistory) // history
			})
		})
	})

	log.Fatal(http.ListenAndServe(":8080", router))
}
