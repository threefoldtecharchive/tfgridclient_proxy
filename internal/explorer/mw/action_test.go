package mw

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type errType struct {
	Error string `json:"error"`
}

var (
	HeaderExample1 = http.Header{
		"C": []string{"D", "E"},
	}
	HeaderExample2 = http.Header{
		"F": []string{"G", "H"},
	}
	HeaderExample3 = http.Header{
		"I": []string{"J", "K"},
	}

	ResponseTextExample1 = "Hello world"
	ResponseTextExample2 = "Hello world2"
	ResponseTextExample3 = "Hello world3"

	ErrExample1 = errors.New("internal grid proxy failure")
	ErrEaxmple2 = errors.New("another internal grid proxy failure")

	JSONErrExample1 = errType{
		Error: ErrExample1.Error(),
	}
	JSONErrExample2 = errType{
		Error: ErrEaxmple2.Error(),
	}

	ProxyHeaderExample1 = http.Header{
		"L": []string{"M"},
	}
	ProxyHeaderExample2 = http.Header{
		"N": []string{"O"},
	}
)

func proxyFailureHandler(r *http.Request) (*http.Response, Response) {
	resp := Error(ErrExample1)
	for k, v := range ProxyHeaderExample1 {
		resp = resp.WithHeader(k, v[0])
	}
	return nil, resp
}

func upstreamFailureHandler(r *http.Request) (*http.Response, Response) {
	resp := http.Response{
		StatusCode: http.StatusBadRequest,
		Header:     HeaderExample1,
		Body:       io.NopCloser(strings.NewReader(ResponseTextExample1)),
	}
	return &resp, nil
}

func successHandler(r *http.Request) (*http.Response, Response) {
	resp := http.Response{
		StatusCode: http.StatusOK,
		Header:     HeaderExample2,
		Body:       io.NopCloser(strings.NewReader(ResponseTextExample2)),
	}
	return &resp, nil
}

// returns the mw.Response
func bothResponsesPresnet(r *http.Request) (*http.Response, Response) {
	httpResp := http.Response{
		StatusCode: http.StatusOK,
		Header:     HeaderExample3,
		Body:       io.NopCloser(strings.NewReader(ResponseTextExample3)),
	}
	resp := Error(ErrEaxmple2)
	for k, v := range ProxyHeaderExample2 {
		resp = resp.WithHeader(k, v[0])
	}

	return &httpResp, resp
}

func TestProxyGridProxyError(t *testing.T) {
	handler := AsProxyHandlerFunc(proxyFailureHandler)
	w := httptest.NewRecorder()
	handler(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Result().StatusCode != http.StatusInternalServerError {
		t.Fatalf("grid proxy status code mismatch: expected: %d, found: %d", http.StatusInternalServerError, w.Result().StatusCode)
	}
	if !reflect.DeepEqual(w.Header(), ProxyHeaderExample1) {
		t.Fatalf("grid proxy header mismatch: expected: %v, found: %v", ProxyHeaderExample1, w.Header())
	}
	var err errType
	if err := json.NewDecoder(w.Body).Decode(&err); err != nil {
		t.Fatalf("failed to decode response body: %s", err.Error())
	}
	if !reflect.DeepEqual(err, JSONErrExample1) {
		t.Fatalf("grid proxy error mismatch: expected: %v, found: %v", JSONErrExample1, err)
	}
}

func TestProxyUpstreamError(t *testing.T) {
	handler := AsProxyHandlerFunc(upstreamFailureHandler)
	w := httptest.NewRecorder()
	handler(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("upstream error status code mismatch: expected: %d, found: %d", http.StatusBadRequest, w.Result().StatusCode)
	}
	if !reflect.DeepEqual(w.Header(), HeaderExample1) {
		t.Fatalf("upstream error header mismatch: expected: %v, found: %v", HeaderExample1, w.Header())
	}
	body := w.Body.String()
	if body != ResponseTextExample1 {
		t.Fatalf("upstream error error mismatch: expected: %v, found: %v", ResponseTextExample1, body)
	}
}

func TestProxySuccess(t *testing.T) {
	handler := AsProxyHandlerFunc(successHandler)
	w := httptest.NewRecorder()
	handler(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("upstream success status code mismatch: expected: %d, found: %d", http.StatusOK, w.Result().StatusCode)
	}
	if !reflect.DeepEqual(w.Header(), HeaderExample2) {
		t.Fatalf("upstream success header mismatch: expected: %v, found: %v", HeaderExample2, w.Header())
	}
	body := w.Body.String()
	if body != ResponseTextExample2 {
		t.Fatalf("upstream success error mismatch: expected: %v, found: %v", ResponseTextExample2, body)
	}
}

func TestBothResponses(t *testing.T) {
	handler := AsProxyHandlerFunc(bothResponsesPresnet)
	w := httptest.NewRecorder()
	handler(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Result().StatusCode != http.StatusInternalServerError {
		t.Fatalf("both result status code mismatch: expected: %d, found: %d", http.StatusInternalServerError, w.Result().StatusCode)
	}
	if !reflect.DeepEqual(w.Header(), ProxyHeaderExample2) {
		t.Fatalf("both result header mismatch: expected: %v, found: %v", ProxyHeaderExample2, w.Header())
	}
	var err errType
	if err := json.NewDecoder(w.Body).Decode(&err); err != nil {
		t.Fatalf("failed to decode response body: %s", err.Error())
	}
	if !reflect.DeepEqual(err, JSONErrExample2) {
		t.Fatalf("both result error mismatch: expected: %v, found: %v", JSONErrExample2, err)
	}
}
