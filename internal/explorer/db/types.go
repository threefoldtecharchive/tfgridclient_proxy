package db

import "github.com/threefoldtech/grid_proxy_server/pkg/types"

// Database interface for storing and fetching grid info
type Database interface {
	GetCounters(filter types.StatsFilter) (types.Counters, error)
	CountNodes() (uint, error)
	GetNode(nodeID uint32) (DBNodeData, error)
	GetFarm(farmID uint32) (types.Farm, error)
	GetNodes(filter types.NodeFilter, limit types.Limit) ([]DBNodeData, uint, error)
	GetFarms(filter types.FarmFilter, limit types.Limit) ([]types.Farm, uint, error)
	GetTwins(filter types.TwinFilter, limit types.Limit) ([]types.Twin, uint, error)
	GetContracts(filter types.ContractFilter, limit types.Limit) ([]DBContract, uint, error)
}

// Contract is contract info
type DBContract struct {
	ContractID        uint                    `json:"contractId"`
	TwinID            uint                    `json:"twinId"`
	State             string                  `json:"state"`
	CreatedAt         uint                    `json:"created_at"`
	Name              string                  `json:"name"`
	NodeID            uint                    `json:"nodeId"`
	DeploymentData    string                  `json:"deployment_data"`
	DeploymentHash    string                  `json:"deployment_hash"`
	NumberOfPublicIps uint                    `json:"number_of_public_ips"`
	Type              string                  `json:"type"`
	ContractBillings  []types.ContractBilling `json:"contract_billings"`
}

// DBNodeData data about nodes which is calculated from the chain
type DBNodeData struct {
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
	UpdatedAt         int64              `json:"updatedAt"`
	CertificationType string             `json:"certificationType"`
	TotalResources    types.Capacity     `json:"total_resources"`
	UsedResources     types.Capacity     `json:"used_resources"`
	PublicConfig      types.PublicConfig `json:"publicConfig"`
	Status            string             `json:"status"` // added node status field for up or down
	Dedicated         bool               `json:"dedicated"`
	RentContractID    uint               `json:"rentContractId"`
	RentedByTwinID    uint               `json:"rentedByTwinId"`
}
