package mw

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type errType struct {
	Status  string `json:",omitempty"`
	Message string `json:",omitempty"`
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

	ExampleResponseText = "Hello world"

	ErrExample1 = errors.New("internal grid proxy failure")
	ErrEaxmple2 = errors.New("another internal grid proxy failure")

	JSONErrExample1 = errType{
		Status:  http.StatusText(http.StatusInternalServerError),
		Message: ErrExample1.Error(),
	}
	JSONErrExample2 = errType{
		Status:  http.StatusText(http.StatusInternalServerError),
		Message: ErrEaxmple2.Error(),
	}

	ProxyHeaderExample1 = http.Header{
		"L": []string{"M"},
	}
	ProxyHeaderExample2 = http.Header{
		"N": []string{"O"},
	}
)

func proxyFailureHandler(r *http.Request) (interface{}, Response) {
	resp := Error(ErrExample1)
	for k, v := range ProxyHeaderExample1 {
		resp = resp.WithHeader(k, v[0])
	}
	return nil, resp
}

func successHandler(r *http.Request) (interface{}, Response) {
	// should return a valid serializable json
	return ExampleResponseText, nil
}

// returns both a valid object and a response
func bothResultsPresent(r *http.Request) (interface{}, Response) {
	obj := map[string]string{"test": "test"}
	resp := Error(ErrEaxmple2)
	for k, v := range ProxyHeaderExample2 {
		resp = resp.WithHeader(k, v[0])
	}

	return obj, resp
}

func TestProxyGridProxyError(t *testing.T) {
	handler := AsProxyHandlerFunc(proxyFailureHandler)
	w := httptest.NewRecorder()
	handler(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Result().StatusCode != http.StatusInternalServerError {
		t.Fatalf("grid proxy status code mismatch: expected: %d, found: %d", http.StatusInternalServerError, w.Result().StatusCode)
	}
	header := w.Header()
	if header["Access-Control-Allow-Origin"][0] != "*" {
		t.Fatalf("invalid Access-Control-Allow-Origin header: %+v", header)
	}
	if header["Content-Type"][0] != "application/json" {
		t.Fatalf("invalid Content-Type header: %+v", header)
	}
	delete(header, "Access-Control-Allow-Origin")
	delete(header, "Content-Type")
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

func TestProxySuccess(t *testing.T) {
	handler := AsProxyHandlerFunc(successHandler)
	w := httptest.NewRecorder()
	handler(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("upstream success status code mismatch: expected: %d, found: %d", http.StatusOK, w.Result().StatusCode)
	}
	body := w.Body
	// response should be json
	var responseText string
	err := json.Unmarshal(body.Bytes(), &responseText)
	if err != nil {
		t.Fatalf("cannot decode upstream result of %v (must be a valid json)", body.String())
	}
	if responseText != ExampleResponseText {
		t.Fatalf("upstream success error mismatch: expected: %v, found: %v", ExampleResponseText, body)
	}
}

func TestBothResults(t *testing.T) {
	handler := AsProxyHandlerFunc(bothResultsPresent)
	w := httptest.NewRecorder()
	handler(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Result().StatusCode != http.StatusInternalServerError {
		t.Fatalf("both result status code mismatch: expected: %d, found: %d", http.StatusInternalServerError, w.Result().StatusCode)
	}
	header := w.Header()
	if header["Access-Control-Allow-Origin"][0] != "*" {
		t.Fatalf("invalid Access-Control-Allow-Origin header: %+v", header)
	}
	if header["Content-Type"][0] != "application/json" {
		t.Fatalf("invalid Content-Type header: %+v", header)
	}
	delete(header, "Access-Control-Allow-Origin")
	delete(header, "Content-Type")
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
