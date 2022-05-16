package explorer

import (
	"github.com/threefoldtech/grid_proxy_server/internal/explorer/db"
	"github.com/threefoldtech/grid_proxy_server/pkg/types"
)

func nodeFromDBNode(info db.DBNodeData) types.Node {
	return types.Node{
		ID:              info.ID,
		NodeID:          info.NodeID,
		FarmID:          info.FarmID,
		TwinID:          info.TwinID,
		Country:         info.Country,
		GridVersion:     info.GridVersion,
		City:            info.City,
		Uptime:          info.Uptime,
		Created:         info.Created,
		FarmingPolicyID: info.FarmingPolicyID,
		UpdatedAt:       info.UpdatedAt,
		TotalResources:  info.TotalResources,
		UsedResources:   info.UsedResources,
		Location: types.Location{
			Country: info.Country,
			City:    info.City,
		},
		PublicConfig:      info.PublicConfig,
		Status:            info.Status,
		CertificationType: info.CertificationType,
		Dedicated:         info.Dedicated,
		RentContractID:    info.RentContractID,
		RentedByTwinID:    info.RentedByTwinID,
	}

}

func nodeWithNestedCapacityFromDBNode(info db.DBNodeData) types.NodeWithNestedCapacity {
	return types.NodeWithNestedCapacity{
		ID:              info.ID,
		NodeID:          info.NodeID,
		FarmID:          info.FarmID,
		TwinID:          info.TwinID,
		Country:         info.Country,
		GridVersion:     info.GridVersion,
		City:            info.City,
		Uptime:          info.Uptime,
		Created:         info.Created,
		FarmingPolicyID: info.FarmingPolicyID,
		UpdatedAt:       info.UpdatedAt,
		Capacity: types.CapacityResult{
			Total: info.TotalResources,
			Used:  info.UsedResources,
		},
		Location: types.Location{
			Country: info.Country,
			City:    info.City,
		},
		PublicConfig:      info.PublicConfig,
		Status:            info.Status,
		CertificationType: info.CertificationType,
		Dedicated:         info.Dedicated,
		RentContractID:    info.RentContractID,
		RentedByTwinID:    info.RentedByTwinID,
	}

}

func contractFromDBContract(info db.DBContract) types.Contract {
	var details interface{}
	switch info.Type {
	case "node":
		details = types.NodeContractDetails{
			NodeID:            info.NodeID,
			DeploymentData:    info.DeploymentData,
			DeploymentHash:    info.DeploymentHash,
			NumberOfPublicIps: info.NumberOfPublicIps,
		}
	case "name":
		details = types.NameContractDetails{
			Name: info.Name,
		}
	case "rent":
		details = types.RentContractDetails{
			NodeID: info.NodeID,
		}
	}
	return types.Contract{
		ContractID: info.ContractID,
		TwinID:     info.TwinID,
		State:      info.State,
		CreatedAt:  info.CreatedAt,
		Type:       info.Type,
		Details:    details,
		Billing:    info.ContractBillings,
	}
}
