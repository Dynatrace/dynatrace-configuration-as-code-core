package rest

import "net/http"

// RequestRetrier represents a component for retrying failed HTTP requests.
type RequestRetrier struct {
	MaxRetries      int
	ShouldRetryFunc func(resp *http.Response) bool
}

// RetryIfNotSuccess implements a basic retry function for a RequestRetrier which will retry on any non 2xx status code
func RetryIfNotSuccess(resp *http.Response) bool {
	return !(resp.StatusCode >= 200 && resp.StatusCode <= 299)
}
