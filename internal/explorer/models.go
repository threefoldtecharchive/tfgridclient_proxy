package explorer

import (
	"encoding/json"

	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/threefoldtech/grid_proxy_server/internal/explorer/db"
	"github.com/threefoldtech/zos/pkg/gridtypes"
	"github.com/threefoldtech/zos/pkg/rmb"
)

// ErrNodeNotFound creates new error type to define node existence or server problem
var (
	ErrNodeNotFound = errors.New("node not found")
)

// ErrBadGateway creates new error type to define node existence or server problem
var (
	ErrBadGateway = errors.New("bad gateway")
	ErrBadRequest = errors.New("bad request")
)

// App is the main app objects
type App struct {
	db             db.Database
	rmb            rmb.Client
	lruCache       *cache.Cache
	releaseVersion string
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

// ErrorReply when something bad happens at grid proxy
type ErrorReply struct {
	Error   string
	Message string
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

type ConnectionInfo struct {
	LastFetchAttempt uint64 `json:"lastFetchAttempt"`
	LastNodeError    string `json:"lastNodeError"`
	Retries          uint64 `json:"retries"`
	ProxyUpdateAt    uint64 `json:"proxyUpdatedAt"`
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
	ConnectionInfo    ConnectionInfo     `json:"connectionInfo"`
}

func nodeFromDBNode(info db.AllNodeData) node {
	return node{
		Version:         info.Graphql.Version,
		ID:              info.Graphql.ID,
		NodeID:          info.NodeID,
		FarmID:          info.Graphql.FarmID,
		TwinID:          info.Graphql.TwinID,
		Country:         info.Graphql.Country,
		GridVersion:     info.Graphql.GridVersion,
		City:            info.Graphql.City,
		Uptime:          info.Graphql.Uptime,
		Created:         info.Graphql.Created,
		FarmingPolicyID: info.Graphql.FarmingPolicyID,
		UpdatedAt:       info.Graphql.UpdatedAt,
		TotalResources:  info.Node.TotalResources,
		UsedResources:   info.Node.UsedResources,
		Location: location{
			Country: info.Graphql.Country,
			City:    info.Graphql.City,
		},
		PublicConfig: publicConfig{
			Domain: info.Graphql.PublicConfig.Domain,
			Gw4:    info.Graphql.PublicConfig.Gw4,
			Gw6:    info.Graphql.PublicConfig.Gw6,
			Ipv4:   info.Graphql.PublicConfig.Ipv4,
			Ipv6:   info.Graphql.PublicConfig.Ipv6,
		},
		Status:            info.Node.Status,
		CertificationType: info.Graphql.CertificationType,
		ConnectionInfo: ConnectionInfo{
			info.ConnectionInfo.LastFetchAttempt,
			info.ConnectionInfo.LastNodeError,
			info.ConnectionInfo.Retries,
			info.ConnectionInfo.ProxyUpdateAt,
		},
	}

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

func farmFromDBFarm(info db.Farm) farm {
	res := farm{
		Name:            info.Name,
		FarmID:          info.FarmID,
		TwinID:          info.TwinID,
		Version:         info.Version,
		PricingPolicyID: info.PricingPolicyID,
		StellarAddress:  info.StellarAddress,
	}
	res.PublicIps = make([]publicIP, len(info.PublicIps))
	for idx, ip := range info.PublicIps {
		res.PublicIps[idx] = publicIP{
			ID:         ip.ID,
			IP:         ip.IP,
			FarmID:     ip.FarmID,
			ContractID: ip.ContractID,
			Gateway:    ip.Gateway,
		}
	}
	return res
}

type publicIP struct {
	ID         string `json:"id"`
	IP         string `json:"ip"`
	FarmID     string `json:"farmId"`
	ContractID int    `json:"contractId"`
	Gateway    string `json:"gateway"`
}

type farmData struct {
	Farms []db.Farm `json:"farms"`
}

// FarmResult is to unmarshal json in it
type FarmResult struct {
	Data farmData `json:"data"`
}

type totalCountResponse struct {
	Date totalCountData `json:"data"`
}

type totalCountData struct {
	NodesConnection nodesConnection `json:"nodesConnection"`
}


type nodesConnection struct {
	TotalCount int `json:"totalCount"`
}
