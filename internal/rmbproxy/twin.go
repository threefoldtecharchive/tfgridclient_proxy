package rmbproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/substrate-client"
)

func submitURL(twinIP string) string {
	return fmt.Sprintf("http://%s:8051/zbus-cmd", twinIP)
}

func resultURL(twinIP string) string {
	return fmt.Sprintf("http://%s:8051/zbus-result", twinIP)
}

// NewTwinResolver : create a new substrate resolver
func NewTwinResolver(substrate *substrate.Substrate, rmbTimeout time.Duration) (*TwinExplorerResolver, error) {

	return &TwinExplorerResolver{
		client:     substrate,
		rmbTimeout: rmbTimeout,
	}, nil
}

func (c *twinClient) SubmitMessage(msg bytes.Buffer) (*http.Response, error) {

	resp, err := c.httpClient.Post(submitURL(c.dstIP), "application/json", &msg)

	if err != nil {
		log.Error().Str("dstIP", c.dstIP).Msg(err.Error())
		return nil, err
	}

	log.Debug().Str("dstIP", c.dstIP).Str("response_status", resp.Status).Msg("Message submitted")
	return resp, nil
}

func (c *twinClient) GetResult(msgIdentifier MessageIdentifier) (*http.Response, error) {

	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(msgIdentifier); err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(resultURL(c.dstIP), "application/json", &buffer)
	if err != nil {
		log.Error().Str("dstIP", c.dstIP).Msg(err.Error())
		return nil, err
	}

	log.Debug().Str("dstIP", c.dstIP).Str("response_status", resp.Status).Msg("Message submitted")
	return resp, err
}
