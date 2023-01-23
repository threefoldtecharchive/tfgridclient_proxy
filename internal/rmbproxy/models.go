package rmbproxy

import (
	"bytes"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/substrate-client"
)

// MessageIdentifier to get the specific result
type MessageIdentifier struct {
	Retqueue string `json:"retqueue" example:"xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"`
}

// App is the main app objects
type App struct {
	resolver *TwinExplorerResolver
}

type PingMessage struct {
	Ping string `json:"ping" example:"pong"`
}

// Flags for the App cmd command
type Flags struct {
	Debug        string
	Substrate    string
	Address      string
	Domain       string
	TLSEmail     string
	CA           string
	CertCacheDir string
}

// TwinExplorerResolver is Substrate resolver
type TwinExplorerResolver struct {
	client     *substrate.Substrate
	rmbTimeout time.Duration
}

// NewTwinClient : create new TwinClient
func (t *TwinExplorerResolver) Get(twinID int) (TwinClient, error) {
	log.Debug().Int("twin", twinID).Msg("resolving twin")
	twin, err := t.client.GetTwin(uint32(twinID))
	if err != nil {
		return nil, err
	}
	log.Debug().Str("ip", twin.IP).Msg("resolved twin ip")

	return &twinClient{
		dstIP:   twin.IP,
		timeout: t.rmbTimeout,
	}, nil
}

type twinClient struct {
	dstIP   string
	timeout time.Duration
}

// TwinClient interface
type TwinClient interface {
	SubmitMessage(msg bytes.Buffer) (*http.Response, error)
	GetResult(msgID MessageIdentifier) (*http.Response, error)
}
