package db

import (
	"sync"

	"github.com/threefoldtech/zos/pkg/gridtypes"
)

type ActiveNodes struct {
	mutex sync.RWMutex
	nodes map[uint32]struct{}
}

func NewActiveNodes() ActiveNodes {
	return ActiveNodes{sync.RWMutex{}, make(map[uint32]struct{})}
}

func (a *ActiveNodes) Add(node uint32) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.nodes[node] = struct{}{}
}

func (a *ActiveNodes) Remove(node uint32) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	delete(a.nodes, node)
}

func (a *ActiveNodes) Has(node uint32) bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	_, ok := a.nodes[node]
	return ok
}

type Limit struct {
	Size uint64
	Page uint64
}

type NodeFilter struct {
	Status   *string
	FreeCRU  *uint64
	FreeMRU  *uint64
	FreeHRU  *uint64
	FreeSRU  *uint64
	Country  *string
	City     *string
	FarmName *string
	FarmIDs  []uint64
	FreeIPs  *uint64
	IPv4     *bool
	IPv6     *bool
	Domain   *bool
}

type FarmFilter struct {
	FreeIPs         *uint64
	StellarAddress  *string
	PricingPolicyID *uint64
	Version         *uint64
	FarmID          *uint64
	TwinID          *uint64
	Name            *string
}

type PublicConfig struct {
	Domain string `json:"domain"`
	Gw4    string `json:"gw4"`
	Gw6    string `json:"gw6"`
	Ipv4   string `json:"ipv4"`
	Ipv6   string `json:"ipv6"`
}

type GraphqlData struct {
	Version           int          `json:"version"`
	ID                string       `json:"id"`
	FarmID            int          `json:"farmId"`
	NodeID            int          `json:"nodeId"`
	TwinID            int          `json:"twinId"`
	Country           string       `json:"country"`
	GridVersion       int          `json:"gridVersion"`
	City              string       `json:"city"`
	Uptime            int64        `json:"uptime"`
	Created           int64        `json:"created"`
	FarmingPolicyID   int          `json:"farmingPolicyId"`
	UpdatedAt         string       `json:"updatedAt"`
	CertificationType string       `json:"certificationType"`
	PublicConfig      PublicConfig `json:"publicConfig"`
}

type NodeData struct {
	TotalResources gridtypes.Capacity `json:"total_resources"`
	UsedResources  gridtypes.Capacity `json:"used_resources"`
	Status         string             `json:"status"` // added node status field for up or down
	Hypervisor     string             `json:"hypervisor"`
	ZosVersion     string             `json:"zosVersion"`
}

type Farm struct {
	Name            string     `json:"name"`
	FarmID          int        `json:"farmId"`
	TwinID          int        `json:"twinId"`
	Version         int        `json:"version"`
	PricingPolicyID int        `json:"pricingPolicyId"`
	StellarAddress  string     `json:"stellarAddress"`
	PublicIps       []PublicIP `json:"publicIps"`
}

type PublicIP struct {
	ID         string `json:"id"`
	IP         string `json:"ip"`
	FarmID     string `json:"farmId"`
	ContractID int    `json:"contractId"`
	Gateway    string `json:"gateway"`
}

type ConnectionInfo struct {
	ProxyUpdateAt    uint64
	LastNodeError    string
	LastFetchAttempt uint64
	Retries          uint64
}

type AllNodeData struct {
	NodeID         int `json:"nodeId"`
	Graphql        GraphqlData
	Node           NodeData
	ConnectionInfo ConnectionInfo
}

type Database interface {
	// if doesn't exist, insert the graphql info
	// otherwise, update the graphql fields
	InsertOrUpdateNodeGraphqlData(nodeID uint32, nodeInfo GraphqlData) error
	// update nongraphql data
	UpdateNodeData(nodeID uint32, nodeInfo NodeData) error
	UpdateNodeError(nodeID uint32, err error) error
	UpdateFarm(farmInfo Farm) error
	GetNode(nodeID uint32) (AllNodeData, error)
	GetFarm(farmID uint32) (Farm, error)
	GetNodes(filter NodeFilter, limit Limit) ([]AllNodeData, error)
	GetFarms(filter FarmFilter, limit Limit) ([]Farm, error)
}

type NodeCursor struct {
	db       Database
	current  int
	pageSize int
}

func NewNodeCursor(db Database, pageSize int) NodeCursor {
	return NodeCursor{db, 1, pageSize}
}

func (nc *NodeCursor) Next() ([]AllNodeData, error) {
	nodes, err := nc.db.GetNodes(NodeFilter{}, Limit{
		Size: uint64(nc.pageSize),
		Page: uint64(nc.current),
	})
	if err != nil {
		return nil, err
	}
	nc.current += 1
	return nodes, nil
}
