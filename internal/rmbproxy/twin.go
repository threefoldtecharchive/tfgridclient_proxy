package rmbproxy

import (
	"bytes"
	"context"
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
func NewTwinResolver(substrate *substrate.Substrate, rmbTimeout int) (*TwinExplorerResolver, error) {

	return &TwinExplorerResolver{
		client:     substrate,
		rmbTimeout: rmbTimeout,
	}, nil
}

func (c *twinClient) SubmitMessage(msg bytes.Buffer) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.timeout))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, submitURL(c.dstIP), &msg)
	if err != nil {
		log.Error().Str("dstIP", c.dstIP).Msg(err.Error())
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	// check on response for non-communication errors?
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.timeout))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resultURL(c.dstIP), &buffer)
	if err != nil {
		log.Error().Str("dstIP", c.dstIP).Msg(err.Error())
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error().Str("dstIP", c.dstIP).Msg(err.Error())
		return nil, err
	}

	log.Debug().Str("dstIP", c.dstIP).Str("response_status", resp.Status).Msg("Message submitted")
	return resp, err
}
