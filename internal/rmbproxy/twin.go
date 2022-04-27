package rmbproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/threefoldtech/substrate-client"
)

func submitURL(twinIP string) string {
	return fmt.Sprintf("http://%s:8051/zbus-cmd", twinIP)
}

func resultURL(twinIP string) string {
	return fmt.Sprintf("http://%s:8051/zbus-result", twinIP)
}

// NewTwinResolver : create a new substrate resolver
func NewTwinResolver(substrateURL string) (*TwinExplorerResolver, error) {
	client, err := substrate.NewSubstrate(substrateURL)
	if err != nil {
		return nil, err
	}

	return &TwinExplorerResolver{
		client: client,
	}, nil
}

func (c *twinClient) readError(r io.Reader) string {
	var body struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r).Decode(&body); err != nil {
		return fmt.Sprintf("failed to read response body: %s", err)
	}

	return body.Message
}

func (c *twinClient) SubmitMessage(msg bytes.Buffer) (*http.Response, error) {
	resp, err := http.Post(submitURL(c.dstIP), "application/json", &msg)
	// check on response for non-communication errors?
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *twinClient) GetResult(msgIdentifier MessageIdentifier) (*http.Response, error) {
	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(msgIdentifier); err != nil {
		return nil, err
	}
	resp, err := http.Post(resultURL(c.dstIP), "application/json", &buffer)

	if err != nil {
		return nil, err
	}

	return resp, err
}
