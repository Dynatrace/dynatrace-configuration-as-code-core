package rest

import (
	"context"
	"net/http"
	"time"
)

// RequestResponseRecorder is a component that records HTTP requests and responses.
type RequestResponseRecorder struct {
	Channel chan RequestResponse
}

// RequestResponse represents a recorded HTTP request and response.
type RequestResponse struct {
	Timestamp time.Time      // Timestamp of the recorded request/response
	Request   *http.Request  // HTTP request
	Response  *http.Response // HTTP response
	Error     error          // Error, if any, during request/response
}

// NewRequestResponseRecorder creates a new RequestResponseRecorder.
func NewRequestResponseRecorder() *RequestResponseRecorder {
	return &RequestResponseRecorder{
		Channel: make(chan RequestResponse),
	}
}

// RecordRequest records an HTTP request.
func (r *RequestResponseRecorder) RecordRequest(ctx context.Context, req *http.Request) {
	if r != nil {
		r.record(ctx, RequestResponse{Timestamp: time.Now(), Request: req})
	}
}

// RecordResponse records an HTTP response or an error.
func (r *RequestResponseRecorder) RecordResponse(ctx context.Context, resp *http.Response, err error) {
	if r != nil {
		reqResp := RequestResponse{Timestamp: time.Now(), Response: resp, Error: err}
		r.record(ctx, reqResp)
	}
}

// record records the RequestResponse values
func (r *RequestResponseRecorder) record(ctx context.Context, reqResp RequestResponse) {
	select {
	case <-ctx.Done():
		return
	case r.Channel <- reqResp:
	}

}

// CancelListening cancels the listening to RequestResponse records.
func (r *RequestResponseRecorder) CancelListening() {
	close(r.Channel)
}
