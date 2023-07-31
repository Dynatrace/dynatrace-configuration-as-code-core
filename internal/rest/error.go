package rest

import "fmt"

// HTTPError represents a custom error type containing the actual HTTP response code and payload.
type HTTPError struct {
	Code    int
	Payload []byte
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("HTTP request failed with status code %d", e.Code)
}
