package rmbproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/threefoldtech/zos/pkg/substrate"
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

func (c *twinClient) SubmitMessage(msg bytes.Buffer) (string, error) {
	resp, err := http.Post(submitURL(c.dstIP), "application/json", &msg)
	// check on response for non-communication errors?
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to Submit the message: %s (%s)", resp.Status, c.readError(resp.Body))
	}

	buffer := new(bytes.Buffer)
	buffer.ReadFrom(resp.Body)
	response := buffer.String()

	return response, nil
}

func (c *twinClient) GetResult(msgIdentifier MessageIdentifier) (string, error) {
	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(msgIdentifier); err != nil {
		return "", err
	}
	resp, err := http.Post(resultURL(c.dstIP), "application/json", &buffer)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// body
		return "", fmt.Errorf("failed to send remote: %s (%s)", resp.Status, c.readError(resp.Body))
	}

	buffer.ReadFrom(resp.Body)
	response := buffer.String()

	return response, err
}
