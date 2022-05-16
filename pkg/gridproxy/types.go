package gridproxy

import (
	"github.com/threefoldtech/zos/pkg/gridtypes"
)

const NodeUP = "up"
const NodeDOWN = "down"

// capacityResult is the NodeData capacity results to unmarshal json in it
type CapacityResult struct {
	Total gridtypes.Capacity `json:"total_resources"`
	Used  gridtypes.Capacity `json:"used_resources"`
}

// NodeInfo is node specific info, queried directly from the node
type NodeInfo struct {
	FarmID       int            `json:"farmId"`
	PublicConfig PublicConfig   `json:"publicConfig"`
	Status       string         `json:"status"` // added node status field for up or down
	Capacity     CapacityResult `json:"capacity"`
}

type PublicConfig struct {
	Domain string `json:"domain"`
	Gw4    string `json:"gw4"`
	Gw6    string `json:"gw6"`
	Ipv4   string `json:"ipv4"`
	Ipv6   string `json:"ipv6"`
}

type ErrorReply struct {
	Error string `json:"error"`
}

// Node is a struct holding the data for a node for the nodes view
type Node struct {
	ID                string       `json:"id"`
	NodeID            uint32       `json:"nodeId"`
	FarmID            int          `json:"farmId"`
	TwinID            int          `json:"twinId"`
	Country           string       `json:"country"`
	GridVersion       int          `json:"gridVersion"`
	City              string       `json:"city"`
	Uptime            int64        `json:"uptime"`
	Created           int64        `json:"created"`
	FarmingPolicyID   int          `json:"farmingPolicyId"`
	TotalResources    Capacity     `json:"total_resources"`
	UsedResources     Capacity     `json:"used_resources"`
	Location          Location     `json:"location"`
	PublicConfig      PublicConfig `json:"publicConfig"`
	Status            string       `json:"status"` // added node status field for up or down
	CertificationType string       `json:"certificationType"`
}

//Capacity is the resources needed for workload(cpu, memory, SSD disk, HDD disks)
type Capacity struct {
	CRU uint64         `json:"cru"`
	SRU gridtypes.Unit `json:"sru"`
	HRU gridtypes.Unit `json:"hru"`
	MRU gridtypes.Unit `json:"mru"`
}

type Location struct {
	Country string `json:"country"`
	City    string `json:"city"`
}
type NodeStatus struct {
	Status string `json:"status"`
}

// Nodes is struct for the whole nodes view
type Nodes struct {
	Data []Node `json:"nodes"`
}

// NodeResponseStruct is struct for the whole nodes view
type NodesResponse struct {
	Nodes Nodes `json:"data"`
}

type NodeID struct {
	NodeID uint32 `json:"nodeId"`
}

// nodeIdData is the nodeIdData to unmarshal json in it
type NodeIDData struct {
	NodeResult []NodeID `json:"nodes"`
}

// nodeIdResult is the nodeIdResult  to unmarshal json in it
type NodeIDResult struct {
	Data NodeIDData `json:"data"`
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

type FarmResult = []Farm

// FarmResult is to unmarshal json in it

// Limit used for pagination
type Limit struct {
	Size     uint64
	Page     uint64
	RetCount bool
}

// NodeFilter node filters
type NodeFilter struct {
	Status       *string
	FreeMRU      *uint64
	FreeHRU      *uint64
	FreeSRU      *uint64
	Country      *string
	City         *string
	FarmName     *string
	FarmIDs      []uint64
	FreeIPs      *uint64
	IPv4         *bool
	IPv6         *bool
	Domain       *bool
	Rentable     *bool
	RentedBy     *uint64
	AvailableFor *uint64
}

// FarmFilter farm filters
type FarmFilter struct {
	FreeIPs           *uint64
	TotalIPs          *uint64
	StellarAddress    *string
	PricingPolicyID   *uint64
	Version           *uint64
	FarmID            *uint64
	TwinID            *uint64
	Name              *string
	NameContains      *string
	CertificationType *string
	Dedicated         *bool
}

// ContractFilter contract filters
type ContractFilter struct {
	ContractID        *uint64
	TwinID            *uint64
	NodeID            *uint64
	Type              *string
	State             *string
	Name              *string
	NumberOfPublicIps *uint64
	DeploymentData    *string
	DeploymentHash    *string
}

// ContractBilling is contract billing info
type ContractBilling struct {
	AmountBilled     uint64 `json:"amountBilled"`
	DiscountReceived string `json:"discountReceived"`
	Timestamp        uint64 `json:"timestamp"`
}

// Contract is contract info
type Contract struct {
	ContractID uint              `json:"contractId"`
	TwinID     uint              `json:"twinId"`
	State      string            `json:"state"`
	CreatedAt  uint              `json:"created_at"`
	Type       string            `json:"type"`
	Details    interface{}       `json:"details"`
	Billing    []ContractBilling `json:"billing"`
}

type NodeContractDetails struct {
	NodeID            uint   `json:"nodeId"`
	DeploymentData    string `json:"deployment_data"`
	DeploymentHash    string `json:"deployment_hash"`
	NumberOfPublicIps uint   `json:"number_of_public_ips"`
}

type NameContractDetails struct {
	Name string `json:"name"`
}

type RentContractDetails struct {
	NodeID uint `json:"nodeId"`
}
