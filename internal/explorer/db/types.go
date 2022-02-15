package db

import (
	"github.com/threefoldtech/zos/pkg/gridtypes"
)

// Limit used for pagination
type Limit struct {
	Size uint64
	Page uint64
}

// NodeFilter node filters
type NodeFilter struct {
	Status   *string
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

// FarmFilter farm filters
type FarmFilter struct {
	FreeIPs         *uint64
	StellarAddress  *string
	PricingPolicyID *uint64
	Version         *uint64
	FarmID          *uint64
	TwinID          *uint64
	Name            *string
}

// PublicConfig node public config
type PublicConfig struct {
	Domain string `json:"domain"`
	Gw4    string `json:"gw4"`
	Gw6    string `json:"gw6"`
	Ipv4   string `json:"ipv4"`
	Ipv6   string `json:"ipv6"`
}

// NodeData data about nodes which is calculated from the chain
type NodeData struct {
	Version           int                `json:"version"`
	ID                string             `json:"id"`
	FarmID            int                `json:"farmId"`
	NodeID            int                `json:"nodeId"`
	TwinID            int                `json:"twinId"`
	Country           string             `json:"country"`
	GridVersion       int                `json:"gridVersion"`
	City              string             `json:"city"`
	Uptime            int64              `json:"uptime"`
	Created           int64              `json:"created"`
	FarmingPolicyID   int                `json:"farmingPolicyId"`
	UpdatedAt         string             `json:"updatedAt"`
	CertificationType string             `json:"certificationType"`
	TotalResources    gridtypes.Capacity `json:"total_resources"`
	PublicConfig      PublicConfig       `json:"publicConfig"`
}

// CapacityInfo capacity info about the node

type CapacityInfo struct {
	UsedCRU   uint64         `json:"cru"`
	FreeSRU   gridtypes.Unit `json:"sru"`
	FreeHRU   gridtypes.Unit `json:"hru"`
	FreeMRU   gridtypes.Unit `json:"mru"`
	UsedIPV4U uint64         `json:"ipv4u"`
}

// PulledNodeData data about nodes which is calculated from communicting with the node
type PulledNodeData struct {
	Resources  CapacityInfo `json:"resources_info"`
	Status     string       `json:"status"` // added node status field for up or down
	Hypervisor string       `json:"hypervisor"`
	ZosVersion string       `json:"zosVersion"`
}

// Farm farm info
type Farm struct {
	Name            string     `json:"name"`
	FarmID          int        `json:"farmId"`
	TwinID          int        `json:"twinId"`
	Version         int        `json:"version"`
	PricingPolicyID int        `json:"pricingPolicyId"`
	StellarAddress  string     `json:"stellarAddress"`
	PublicIps       []PublicIP `json:"publicIps"`
}

// PublicIP info about public ip in the farm
type PublicIP struct {
	ID         string `json:"id"`
	IP         string `json:"ip"`
	FarmID     string `json:"farmId"`
	ContractID int    `json:"contractId"`
	Gateway    string `json:"gateway"`
}

// ConnectionInfo info about connections to the nodes
type ConnectionInfo struct {
	ProxyUpdateAt uint64 `json:"proxyUpdatedAt"`
}

// AllNodeData contains info from the chain, the node, connection info
type AllNodeData struct {
	NodeID         int `json:"nodeId"`
	NodeData       NodeData
	PulledNodeData PulledNodeData
	ProxyUpdatedAt uint64
}

// Counters contains aggregate info about the grid
type Counters struct {
	Nodes       uint64 `json:"nodes"`
	Farms       uint64 `json:"farms"`
	Countries   uint64 `json:"countries"`
	TotalCRU    uint64 `json:"totalCru"`
	TotalSRU    uint64 `json:"totalSru"`
	TotalMRU    uint64 `json:"totalMru"`
	TotalHRU    uint64 `json:"totalHru"`
	PublicIPs   uint64 `json:"publicIps"`
	AccessNodes uint64 `json:"accessNodes"`
	Gateways    uint64 `json:"gateways"`
	Twins       uint64 `json:"twins"`
	Contracts   uint64 `json:"contracts"`
}

// Database interface for storing and fetching grid info
type Database interface {
	GetCounters() (Counters, error)
	CountNodes() (int, error)
	UpdateNodeData(nodeID uint32, nodeInfo PulledNodeData) error
	UpdateNodeDataByTwin(twinID uint32, nodeInfo PulledNodeData) error
	GetNode(nodeID uint32) (AllNodeData, error)
	GetFarm(farmID uint32) (Farm, error)
	GetNodes(filter NodeFilter, limit Limit) ([]AllNodeData, error)
	GetFarms(filter FarmFilter, limit Limit) ([]Farm, error)
}

// NodeCursor for pagination
type NodeCursor struct {
	db       Database
	current  int
	pageSize int
}

// NewNodeCursor return a paginator over the db with the given page size
func NewNodeCursor(db Database, pageSize int) NodeCursor {
	return NodeCursor{db, 1, pageSize}
}

// Next returns the next node patch
func (nc *NodeCursor) Next() ([]AllNodeData, error) {
	nodes, err := nc.db.GetNodes(NodeFilter{}, Limit{
		Size: uint64(nc.pageSize),
		Page: uint64(nc.current),
	})
	if err != nil {
		return nil, err
	}
	nc.current++
	return nodes, nil
}
