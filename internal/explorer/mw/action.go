package mw

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

// Response interface
type Response interface {
	Status() int
	Err() error

	// header getter
	Header() http.Header
	// header setter
	WithHeader(k, v string) Response
}

// Action interface
type Action func(r *http.Request) (interface{}, Response)

// ProxyAction interface
type ProxyAction func(r *http.Request) (*http.Response, Response)

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func exposeHeaders(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Expose-Headers", "*")
}

type ProxyResult struct {
	Status  string `json:",omitempty"`
	Message string `json:",omitempty"`
}

// AsProxyHandlerFunc is the same as AsHandlerFunc
// except it returns a different result in case of error response
// this can be modified to support both, but kept as is for now
// to be compatible with the rmb result
func AsProxyHandlerFunc(a Action) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			_, _ = io.ReadAll(r.Body)
			_ = r.Body.Close()
		}()
		enableCors(&w)
		result, response := a(r)

		var headers http.Header
		var statusCode int

		// headers
		if response != nil {
			headers = response.Header()
			headers.Set("Content-Type", "application/json")
		}

		for k, v := range headers {
			// override if present
			w.Header().Set(k, v[0])
			// add all
			for _, v := range v[1:] {
				w.Header().Add(k, v)
			}
		}

		// status code
		if response != nil {
			statusCode = response.Status()
		} else {
			statusCode = http.StatusOK
		}

		w.WriteHeader(statusCode)

		// body
		if response != nil && response.Err() != nil {
			// to be consistent with https://github.com/threefoldtech/rmb_go/blob/825c23c921d395294f3d28d5d9f1d009e8fde9d6/models.go#L35
			object := ProxyResult{
				Message: response.Err().Error(),
				Status:  http.StatusText(response.Status()),
			}
			if err := json.NewEncoder(w).Encode(object); err != nil {
				log.Error().Err(err).Msg("failed to encode return object")
			}
		} else {
			if err := json.NewEncoder(w).Encode(result); err != nil {
				log.Error().Err(err).Msg("failed to encode return object")
			}
		}
	}
}

// AsHandlerFunc is a helper wrapper to make implementing actions easier
func AsHandlerFunc(a Action) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			_, _ = io.ReadAll(r.Body)
			_ = r.Body.Close()
		}()
		enableCors(&w)
		exposeHeaders(&w)

		object, result := a(r)

		w.Header().Set("Content-Type", "application/json")

		if result == nil {
			w.WriteHeader(http.StatusOK)
		} else {

			h := result.Header()
			for k := range h {
				for _, v := range h.Values(k) {
					w.Header().Add(k, v)
				}
			}

			w.WriteHeader(result.Status())
			if err := result.Err(); err != nil {
				log.Error().Msgf("%s", err.Error())
				object = struct {
					Error string `json:"error"`
				}{
					Error: err.Error(),
				}
			}
		}

		if err := json.NewEncoder(w).Encode(object); err != nil {
			log.Error().Err(err).Msg("failed to encode return object")
		}
	}
}

type genericResponse struct {
	status int
	err    error
	header http.Header
}

func (r genericResponse) Status() int {
	return r.status
}

func (r genericResponse) Err() error {
	return r.err
}

func (r genericResponse) Header() http.Header {
	if r.header == nil {
		r.header = http.Header{}
	}
	return r.header
}

func (r genericResponse) WithHeader(k, v string) Response {
	if r.header == nil {
		r.header = http.Header{}
	}

	r.header.Add(k, v)
	return r
}

// Created return a created response
func Created() Response {
	return genericResponse{status: http.StatusCreated}
}

// Ok return a ok response
func Ok() Response {
	return genericResponse{status: http.StatusOK}
}

// Error generic error response
func Error(err error, code ...int) Response {
	status := http.StatusInternalServerError
	if len(code) > 0 {
		status = code[0]
	}

	if err == nil {
		err = fmt.Errorf("no message")
	}

	return genericResponse{status: status, err: err}
}

// BadRequest result
func BadRequest(err error) Response {
	return Error(err, http.StatusBadRequest)
}

// BadGateway result
func BadGateway(err error) Response {
	return Error(err, http.StatusBadGateway)
}

// PaymentRequired result
func PaymentRequired(err error) Response {
	return Error(err, http.StatusPaymentRequired)
}

// NotFound response
func NotFound(err error) Response {
	return Error(err, http.StatusNotFound)
}

// Conflict response
func Conflict(err error) Response {
	return Error(err, http.StatusConflict)
}

// UnAuthorized response
func UnAuthorized(err error) Response {
	return Error(err, http.StatusUnauthorized)
}

// Forbidden response
func Forbidden(err error) Response {
	return Error(err, http.StatusForbidden)
}

// NoContent response
func NoContent() Response {
	return genericResponse{status: http.StatusNoContent}
}

// Accepted response
func Accepted() Response {
	return genericResponse{status: http.StatusAccepted}
}

// Unavailable returned when server is too busy
func Unavailable(err error) Response {
	return Error(err, http.StatusServiceUnavailable)
}
