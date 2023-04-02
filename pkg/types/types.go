package types

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/threefoldtech/zos/pkg/gridtypes"
)

// ContractBilling is contract billing info
type ContractBilling struct {
	AmountBilled     uint64 `json:"amountBilled"`
	DiscountReceived string `json:"discountReceived"`
	Timestamp        uint64 `json:"timestamp"`
}

// Counters contains aggregate info about the grid
type Counters struct {
	Nodes             int64            `json:"nodes"`
	Farms             int64            `json:"farms"`
	Countries         int64            `json:"countries"`
	TotalCRU          int64            `json:"totalCru"`
	TotalSRU          int64            `json:"totalSru"`
	TotalMRU          int64            `json:"totalMru"`
	TotalHRU          int64            `json:"totalHru"`
	PublicIPs         int64            `json:"publicIps"`
	AccessNodes       int64            `json:"accessNodes"`
	Gateways          int64            `json:"gateways"`
	Twins             int64            `json:"twins"`
	Contracts         int64            `json:"contracts"`
	NodesDistribution map[string]int64 `json:"nodesDistribution" gorm:"-:all"`
}

// PublicConfig node public config
type PublicConfig struct {
	Domain string `json:"domain"`
	Gw4    string `json:"gw4"`
	Gw6    string `json:"gw6"`
	Ipv4   string `json:"ipv4"`
	Ipv6   string `json:"ipv6"`
}

// Capacity is the resources needed for workload(cpu, memory, SSD disk, HDD disks)
type Capacity struct {
	CRU uint64         `json:"cru"`
	SRU gridtypes.Unit `json:"sru"`
	HRU gridtypes.Unit `json:"hru"`
	MRU gridtypes.Unit `json:"mru"`
}
type Farm struct {
	Name              string     `json:"name"`
	FarmID            int        `json:"farmId"`
	TwinID            int        `json:"twinId"`
	PricingPolicyID   int        `json:"pricingPolicyId"`
	CertificationType string     `json:"certificationType"`
	StellarAddress    string     `json:"stellarAddress"`
	Dedicated         bool       `json:"dedicated"`
	PublicIps         []PublicIP `json:"publicIps"`
}

// PublicIP info about public ip in the farm
type PublicIP struct {
	ID         string `json:"id"`
	IP         string `json:"ip"`
	FarmID     string `json:"farmId"`
	ContractID int    `json:"contractId"`
	Gateway    string `json:"gateway"`
}

// StatsFilter statistics filters
type StatsFilter struct {
	Status *string
}

// Limit used for pagination
type Limit struct {
	Size      uint64
	Page      uint64
	RetCount  bool
	Randomize bool
}

// NodeFilter node filters
type NodeFilter struct {
	Status            *string
	FreeMRU           *uint64
	FreeHRU           *uint64
	FreeSRU           *uint64
	TotalMRU          *uint64
	TotalHRU          *uint64
	TotalSRU          *uint64
	TotalCRU          *uint64
	Country           *string
	CountryContains   *string
	City              *string
	CityContains      *string
	FarmName          *string
	FarmNameContains  *string
	FarmIDs           []uint64
	FreeIPs           *uint64
	IPv4              *bool
	IPv6              *bool
	Domain            *bool
	Dedicated         *bool
	Rentable          *bool
	Rented            *bool
	RentedBy          *uint64
	AvailableFor      *uint64
	NodeID            *uint64
	TwinID            *uint64
	CertificationType *string
}

// FarmFilter farm filters
type FarmFilter struct {
	FreeIPs           *uint64
	TotalIPs          *uint64
	StellarAddress    *string
	PricingPolicyID   *uint64
	FarmID            *uint64
	TwinID            *uint64
	Name              *string
	NameContains      *string
	CertificationType *string
	Dedicated         *bool
}

// TwinFilter twin filters
type TwinFilter struct {
	TwinID    *uint64
	AccountID *string
	Relay     *string
	PublicKey *string
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

type Location struct {
	Country   string   `json:"country"`
	City      string   `json:"city"`
	Longitude *float64 `json:"longitude"`
	Latitude  *float64 `json:"latitude"`
}

// Node is a struct holding the data for a Node for the nodes view
type Node struct {
	ID                string       `json:"id"`
	NodeID            int          `json:"nodeId"`
	FarmID            int          `json:"farmId"`
	TwinID            int          `json:"twinId"`
	Country           string       `json:"country"`
	GridVersion       int          `json:"gridVersion"`
	City              string       `json:"city"`
	Uptime            int64        `json:"uptime"`
	Created           int64        `json:"created"`
	FarmingPolicyID   int          `json:"farmingPolicyId"`
	UpdatedAt         int64        `json:"updatedAt"`
	TotalResources    Capacity     `json:"total_resources"`
	UsedResources     Capacity     `json:"used_resources"`
	Location          Location     `json:"location"`
	PublicConfig      PublicConfig `json:"publicConfig"`
	Status            string       `json:"status"` // added node status field for up or down
	CertificationType string       `json:"certificationType"`
	Dedicated         bool         `json:"dedicated"`
	RentContractID    uint         `json:"rentContractId"`
	RentedByTwinID    uint         `json:"rentedByTwinId"`
	SerialNumber      string       `json:"serialNumber"`
}

// CapacityResult is the NodeData capacity results to unmarshal json in it
type CapacityResult struct {
	Total Capacity `json:"total_resources"`
	Used  Capacity `json:"used_resources"`
}

// Node to be compatible with old view
type NodeWithNestedCapacity struct {
	ID                string         `json:"id"`
	NodeID            int            `json:"nodeId"`
	FarmID            int            `json:"farmId"`
	TwinID            int            `json:"twinId"`
	Country           string         `json:"country"`
	GridVersion       int            `json:"gridVersion"`
	City              string         `json:"city"`
	Uptime            int64          `json:"uptime"`
	Created           int64          `json:"created"`
	FarmingPolicyID   int            `json:"farmingPolicyId"`
	UpdatedAt         int64          `json:"updatedAt"`
	Capacity          CapacityResult `json:"capacity"`
	Location          Location       `json:"location"`
	PublicConfig      PublicConfig   `json:"publicConfig"`
	Status            string         `json:"status"` // added node status field for up or down
	CertificationType string         `json:"certificationType"`
	Dedicated         bool           `json:"dedicated"`
	RentContractID    uint           `json:"rentContractId"`
	RentedByTwinID    uint           `json:"rentedByTwinId"`
	SerialNumber      string         `json:"serialNumber"`
}

type Twin struct {
	TwinID    uint   `json:"twinId"`
	AccountID string `json:"accountId"`
	Relay     string `json:"relay"`
	PublicKey string `json:"publicKey"`
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

type Contract struct {
	ContractID uint              `json:"contractId"`
	TwinID     uint              `json:"twinId"`
	State      string            `json:"state"`
	CreatedAt  uint              `json:"created_at"`
	Type       string            `json:"type"`
	Details    interface{}       `json:"details"`
	Billing    []ContractBilling `json:"billing"`
}

type Version struct {
	Version string `json:"version"`
}

type NodeStatisticsResources struct {
	CRU   int `json:"cru"`
	HRU   int `json:"hru"`
	IPV4U int `json:"ipv4u"`
	MRU   int `json:"mru"`
	SRU   int `json:"sru"`
}

type NodeStatisticsUsers struct {
	Deployments int `json:"deployments"`
	Workloads   int `json:"workloads"`
}

type NodeStatistics struct {
	System NodeStatisticsResources `json:"system"`
	Total  NodeStatisticsResources `json:"total"`
	Used   NodeStatisticsResources `json:"used"`
	Users  NodeStatisticsUsers     `json:"users"`
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
