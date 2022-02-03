package explorer

import (
	"encoding/json"

	"github.com/gomodule/redigo/redis"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/threefoldtech/zos/pkg/gridtypes"
	"github.com/threefoldtech/zos/pkg/rmb"
)

// DefaultExplorerURL is the default explorer graphql url
const DefaultExplorerURL string = "https://graphql.dev.grid.tf/graphql"

// ErrNodeNotFound creates new error type to define node existence or server problem
var (
	ErrNodeNotFound = errors.New("node not found")
)

// ErrBadGateway creates new error type to define node existence or server problem
var (
	ErrBadGateway = errors.New("bad gateway")
)

// App is the main app objects
type App struct {
	explorer       string
	redis          *redis.Pool
	rmb            rmb.Client
	lruCache       *cache.Cache
	releaseVersion string
}

// OffsetKey is the type holds the request context
type offsetKey struct{}

// SpecificFarmKey is the type holds the request context
type specificFarmKey struct{}

// MaxResultKey is the type holds the request context
type maxResultKey struct{}

// isGatewayKey is the type holds the request context
type isGatewayKey struct{}

// NodeTwinID is the node twin ID to unmarshal json in it
type nodeTwinID struct {
	TwinID uint32 `json:"twinId"`
}

// NodeData is having nodeTwinID to unmarshal json in it
type nodeData struct {
	NodeResult []nodeTwinID `json:"nodes"`
}

// NodeResult is the NodeData to unmarshal nodeTwinID json in it
type nodeResult struct {
	Data nodeData `json:"data"`
}

// CapacityResult is the NodeData capacity results to unmarshal json in it
type capacityResult struct {
	Total gridtypes.Capacity `json:"total_resources"`
	Used  gridtypes.Capacity `json:"used_resources"`
}

// NodeInfo is node specific info, queried directly from the node
type NodeInfo struct {
	Capacity   capacityResult `json:"capacity"`
	Hypervisor string         `json:"hypervisor"`
	ZosVersion string         `json:"zosVersion"`
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

// NodeStatus is used for status endpoint to decode json in
type NodeStatus struct {
	Status string `json:"status"`
}

// Serialize is the serializer for node status struct
func (n *NodeStatus) Serialize() (json.RawMessage, error) {
	bytes, err := json.Marshal(n)
	if err != nil {
		return json.RawMessage{}, errors.Wrap(err, "failed to serialize json data for node status struct")
	}
	return json.RawMessage(bytes), nil
}

// Deserialize is the deserializer for node status struct
func (n *NodeStatus) Deserialize(data []byte) error {
	err := json.Unmarshal(data, n)
	if err != nil {
		return errors.Wrap(err, "failed to deserialize json data for node status struct")
	}
	return nil
}

type location struct {
	Country string `json:"country"`
	City    string `json:"city"`
}

type publicConfig struct {
	Domain string `json:"domain"`
	Gw4    string `json:"gw4"`
	Gw6    string `json:"gw6"`
	Ipv4   string `json:"ipv4"`
	Ipv6   string `json:"ipv6"`
}

// Node is a struct holding the data for a node for the nodes view
type node struct {
	Version           int                `json:"version"`
	ID                string             `json:"id"`
	NodeID            int                `json:"nodeId"`
	FarmID            int                `json:"farmId"`
	TwinID            int                `json:"twinId"`
	Country           string             `json:"country"`
	GridVersion       int                `json:"gridVersion"`
	City              string             `json:"city"`
	Uptime            int64              `json:"uptime"`
	Created           int64              `json:"created"`
	FarmingPolicyID   int                `json:"farmingPolicyId"`
	UpdatedAt         string             `json:"updatedAt"`
	TotalResources    gridtypes.Capacity `json:"total_resources"`
	UsedResources     gridtypes.Capacity `json:"used_resources"`
	Location          location           `json:"location"`
	PublicConfig      publicConfig       `json:"publicConfig"`
	Status            string             `json:"status"` // added node status field for up or down
	CertificationType string             `json:"certificationType"`
}

// Nodes is struct for the whole nodes view
type nodes struct {
	Data []node `json:"nodes"`
}

// NodeResponseStruct is struct for the whole nodes view
type nodesResponse struct {
	Nodes nodes `json:"data"`
}

type nodeID struct {
	NodeID uint32 `json:"nodeId"`
}

// nodeIdData is the nodeIdData to unmarshal json in it
type nodeIDData struct {
	NodeResult []nodeID `json:"nodes"`
}

// nodeIdResult is the nodeIdResult  to unmarshal json in it
type nodeIDResult struct {
	Data nodeIDData `json:"data"`
}

type farm struct {
	Name            string     `json:"name"`
	FarmID          int        `json:"farmId"`
	TwinID          int        `json:"twinId"`
	Version         int        `json:"version"`
	PricingPolicyID int        `json:"pricingPolicyId"`
	StellarAddress  string     `json:"stellarAddress"`
	PublicIps       []publicIP `json:"publicIps"`
}

type publicIP struct {
	ID         string `json:"id"`
	IP         string `json:"ip"`
	FarmID     string `json:"farmId"`
	ContractID int    `json:"contractId"`
	Gateway    string `json:"gateway"`
}

type farmData struct {
	Farms []farm `json:"farms"`
}

// FarmResult is to unmarshal json in it
type FarmResult struct {
	Data farmData `json:"data"`
}
