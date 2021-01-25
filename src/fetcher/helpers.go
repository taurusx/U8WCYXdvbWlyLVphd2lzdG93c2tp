package main

import (
	"errors"
	"fmt"
	"math"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
)

func getIDParam(r *http.Request) (uint, error) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)

	var sb strings.Builder

	if err != nil {
		sb.WriteString(fmt.Sprintf("Incorrect id format: %v\n", id))
	} else if id < 0 {
		sb.WriteString(fmt.Sprintf("Incorrect id: %v, must be above 0\n", id))
	} else if id == 0 {
		sb.WriteString(fmt.Sprintf("Incorrect id: %v, must be positive value\n", id))
	}

	if sb.Len() == 0 {
		return uint(id), nil
	}

	return uint(id), errors.New(sb.String())
}

// HasContentType determines whether the request `content-type` includes a
// server-acceptable mime-type
//
// Failure should yield an HTTP 415 (`http.StatusUnsupportedMediaType`)
// https://gist.github.com/rjz/fe283b02cbaa50c5991e1ba921adf7c9
func HasContentType(r *http.Request, mimetype string) bool {
	contentType := r.Header.Get("Content-type")
	if contentType == "" {
		return mimetype == "application/octet-stream"
	}

	for _, v := range strings.Split(contentType, ",") {
		t, _, err := mime.ParseMediaType(v)
		if err != nil {
			break
		}
		if t == mimetype {
			return true
		}
	}
	return false
}

func toFixed(num float64, precision int) float64 {
	r := math.Pow10(precision)
	return math.Round(num*r) / r
}
