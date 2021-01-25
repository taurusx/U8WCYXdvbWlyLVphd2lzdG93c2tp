package main

import (
	"math"
	"time"
)

var nanos = math.Pow10(9)

// FetchLog is a log entry of fetch request that has been made
type FetchLog struct {
	Response  *string `json:"response"`
	Duration  float32 `json:"duration,omitempty"`   // seconds
	CreatedAt float64 `json:"created_at,omitempty"` // CreatedAt is a Unix time - number of seconds that elapsed from Unix epoch
}

func (fl *FetchLog) close(endTime time.Time, response *[]byte) {
	nowSec := float64(endTime.UnixNano()) / nanos
	duration := toFixed(nowSec-fl.CreatedAt, 3)
	fl.Duration = float32(duration)

	if response == nil {
		fl.Response = nil
		return
	}
	bodyS := string(*response)
	fl.Response = &bodyS
}

// NewFetchLog initializes FetchLog with timestamp
func NewFetchLog(initTime time.Time) *FetchLog {
	tickSec := float64(initTime.UnixNano()) / nanos
	createdAt := toFixed(tickSec, 5)

	return &FetchLog{CreatedAt: createdAt}
}
