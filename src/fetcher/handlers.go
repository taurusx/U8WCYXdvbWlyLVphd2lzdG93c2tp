package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
)

// bytes sizes
const (
	_ int64 = 1 << (iota * 10) // 1 B
	KB
	MB
)

type addFetcherResponse struct {
	ID uint `json:"id,omitempty"`
}

type updateFetcherResponse struct {
	ID uint `json:"id,omitempty"`
}

func getFetchers(w http.ResponseWriter, r *http.Request) {
	fetchers.Lock()
	defer fetchers.Unlock()

	m := fetchers.fetchersMap

	// sort before returning
	keys := make([]uint, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	fetchersSlice := make([]*Fetcher, 0, len(m))

	for _, key := range keys {
		fetchersSlice = append(fetchersSlice, m[key])
	}

	json.NewEncoder(w).Encode(fetchersSlice)
}

func getFetcher(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fetchers.Lock()
	defer fetchers.Unlock()
	if f, ok := fetchers.fetchersMap[id]; ok {
		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(f)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func addFetcher(w http.ResponseWriter, r *http.Request) {
	// TODO: uncomment to block invalid mime types
	// if ok := HasContentType(r, "application/json"); !ok {
	// 	http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
	// 	return
	// }

	if r.ContentLength > MB {
		http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
		return
	}

	var fetcher Fetcher
	err := json.NewDecoder(r.Body).Decode(&fetcher)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fetchers.Lock()
	defer fetchers.Unlock()
	fetcher.ID = fetchers.nextIndex

	err = fetcher.validate()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fetchers.nextIndex = fetchers.nextIndex + 1
	fetchers.fetchersMap[fetcher.ID] = &fetcher

	fetcher.initWorker()

	newID := addFetcherResponse{ID: fetcher.ID}
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newID)
}

func updateFetcher(w http.ResponseWriter, r *http.Request) {
	// TODO: uncomment to block invalid mime types
	// if ok := HasContentType(r, "application/json"); !ok {
	// 	http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
	// 	return
	// }

	if r.ContentLength > MB {
		http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
		return
	}

	var fetcher Fetcher
	err := json.NewDecoder(r.Body).Decode(&fetcher)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = fetcher.validate()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fetchers.Lock()
	defer fetchers.Unlock()

	if f, ok := fetchers.fetchersMap[fetcher.ID]; ok {
		fetcher.History = f.History
		f.quit <- struct{}{}

		fetchers.fetchersMap[fetcher.ID] = &fetcher
		fetcher.initWorker()

		newID := updateFetcherResponse{ID: fetcher.ID}

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(newID)

		return
	}

	http.Error(w, fmt.Sprintf("Fetcher with id: %v does not exist", fetcher.ID), http.StatusBadRequest)
}

func removeFetcher(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fetchers.Lock()
	defer fetchers.Unlock()
	if f, ok := fetchers.fetchersMap[id]; ok {
		removed := f
		delete(fetchers.fetchersMap, id)

		f.quit <- struct{}{}

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(removed)

		return
	}

	http.Error(w, fmt.Sprintf("Cannot remove fetcher with id: %v. Fetcher not found", id), http.StatusBadRequest)
}

func getFetcherHistory(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fetchers.Lock()
	defer fetchers.Unlock()
	if f, ok := fetchers.fetchersMap[id]; ok {
		history := make([]FetchLog, 0, len(f.History))
		for _, fetchRequest := range f.History {
			history = append(history, *fetchRequest)
		}

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(history)

		return
	}

	http.Error(w, fmt.Sprintf("Cannot read history of fetcher with id: %v. Fetcher not found", id), http.StatusBadRequest)
}
