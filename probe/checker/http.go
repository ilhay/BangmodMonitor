package checker

import (
	"net/http"
	"time"
)

type Result struct {
	URL        string
	StatusCode int
	ResponseMS int64
	IsUp       bool
	Error      string
}

var client = &http.Client{
	Timeout: 10 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return http.ErrUseLastResponse
		}
		return nil
	},
}

func CheckHTTP(url string) Result {
	start := time.Now()

	resp, err := client.Get(url)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		return Result{URL: url, ResponseMS: elapsed, IsUp: false, Error: err.Error()}
	}
	defer resp.Body.Close()

	isUp := resp.StatusCode < 500
	return Result{
		URL:        url,
		StatusCode: resp.StatusCode,
		ResponseMS: elapsed,
		IsUp:       isUp,
	}
}
