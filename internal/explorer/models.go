package explorer

import (
	"encoding/json"

	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/threefoldtech/grid_proxy_server/internal/explorer/db"
	"github.com/threefoldtech/grid_proxy_server/pkg/types"
	"github.com/threefoldtech/rmb-sdk-go"
)

// ErrNodeNotFound creates new error type to define node existence or server problem
var (
	ErrNodeNotFound    = errors.New("node not found")
	ErrGatewayNotFound = errors.New("gateway not found")
)

// ErrBadGateway creates new error type to define node existence or server problem
var (
	ErrBadGateway = errors.New("bad gateway")
	ErrBadRequest = errors.New("bad request")
)

// App is the main app objects
type App struct {
	db             db.Database
	lruCache       *cache.Cache
	releaseVersion string
	relayClient    rmb.Client
}

type ErrorMessage struct {
	Message string `json:"message"`
}

// NodeInfo is node specific info, queried directly from the node
type NodeInfo struct {
	Capacity   types.CapacityResult `json:"capacity"`
	Hypervisor string               `json:"hypervisor"`
	ZosVersion string               `json:"zosVersion"`
}

// Serialize is the serializer for node info struct
func (n *NodeInfo) Serialize() (json.RawMessage, error) {
	bytes, err := json.Marshal(n)
	if err != nil {
		return json.RawMessage{}, errors.Wrap(err, "failed to serialize json data for node info struct")
	}
	return json.RawMessage(bytes), nil
}

// Deserialize is the deserializer for node info struct
func (n *NodeInfo) Deserialize(data []byte) error {
	err := json.Unmarshal(data, n)
	if err != nil {
		return errors.Wrap(err, "failed to deserialize json data for node info struct")
	}
	return nil
}

type PingMessage struct {
	Ping string `json:"ping" example:"pong"`
}
