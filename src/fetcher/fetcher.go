package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const timeout time.Duration = 5 * time.Second // 5 sec

// Fetcher is a representation of fetcher data
type Fetcher struct {
	ID       uint          `json:"id,omitempty"`
	URL      string        `json:"url,omitempty"`
	Interval uint32        `json:"interval,omitempty"`
	History  []*FetchLog   `json:"-"`
	quit     chan struct{} `json:"-"`
}

// initWorker initiates data fetching in intervals
func (f *Fetcher) initWorker() {
	fmt.Printf("[%v - Start] %v\n", f.ID, time.Now())
	fetchLog := NewFetchLog(time.Now())
	go f.fetchFromURL(fetchLog)

	ticker := time.NewTicker(time.Duration(f.Interval) * time.Second)
	quit := make(chan struct{})
	f.quit = quit

	go func() {
		for {
			select {
			case t := <-ticker.C:
				fmt.Printf("[%v - Tick] %v\n", f.ID, t)
				fetchLog := NewFetchLog(t)
				go f.fetchFromURL(fetchLog)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (f *Fetcher) fetchFromURL(fetchLog *FetchLog) {
	client := http.Client{Timeout: timeout}
	res, err := client.Get(f.URL)

	if err != nil {
		fmt.Println(err.Error())

		fetchLog.close(time.Now(), nil)
		f.addLog(fetchLog)
		return
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		fmt.Println(err.Error())

		fetchLog.close(time.Now(), nil)
		f.addLog(fetchLog)
		return
	}

	fetchLog.close(time.Now(), &body)
	f.addLog(fetchLog)
}

// addLog adds FetchLog to Fetcher History
func (f *Fetcher) addLog(fetchLog *FetchLog) {
	f.History = append(f.History, fetchLog)
}

// validate informs if the data inside a Fetcher is valid
func (f *Fetcher) validate() error {
	var sb strings.Builder

	if f.ID == 0 {
		sb.WriteString(fmt.Sprintf("Incorrect id: %v, must be positive value\n", f.ID))
	}

	if f.URL == "" {
		sb.WriteString(fmt.Sprintf("Incorrect URL: %q\n", f.URL))
	}

	if f.Interval == 0 {
		sb.WriteString(fmt.Sprintf("Incorrect interval: %v, must be positive value\n", f.Interval))
	}

	if sb.Len() == 0 {
		return nil
	}

	return errors.New(sb.String())
}
