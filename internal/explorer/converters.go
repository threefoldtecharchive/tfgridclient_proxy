package explorer

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/threefoldtech/grid_proxy_server/internal/explorer/db"
	"github.com/threefoldtech/grid_proxy_server/pkg/types"
	"github.com/threefoldtech/zos/pkg/gridtypes"
)

func nodeFromDBNode(info db.Node) types.Node {
	node := types.Node{
		ID:              info.ID,
		NodeID:          int(info.NodeID),
		FarmID:          int(info.FarmID),
		TwinID:          int(info.TwinID),
		Country:         info.Country,
		GridVersion:     int(info.GridVersion),
		City:            info.City,
		Uptime:          info.Uptime,
		Created:         info.Created,
		FarmingPolicyID: int(info.FarmingPolicyID),
		UpdatedAt:       info.UpdatedAt,
		TotalResources: types.Capacity{
			CRU: uint64(info.TotalCru),
			SRU: gridtypes.Unit(info.TotalSru),
			HRU: gridtypes.Unit(info.TotalHru),
			MRU: gridtypes.Unit(info.TotalMru),
		},
		UsedResources: types.Capacity{
			CRU: uint64(info.UsedCru),
			SRU: gridtypes.Unit(info.UsedSru),
			HRU: gridtypes.Unit(info.UsedHru),
			MRU: gridtypes.Unit(info.UsedMru),
		},
		Location: types.Location{
			Country:   info.Country,
			City:      info.City,
			Longitude: info.Longitude,
			Latitude:  info.Latitude,
		},
		PublicConfig: types.PublicConfig{
			Domain: info.Domain,
			Gw4:    info.Gw4,
			Gw6:    info.Gw6,
			Ipv4:   info.Ipv4,
			Ipv6:   info.Ipv6,
		},
		CertificationType: info.Certification,
		Dedicated:         info.Dedicated,
		RentContractID:    uint(info.RentContractID),
		RentedByTwinID:    uint(info.RentedByTwinID),
		SerialNumber:      info.SerialNumber,
	}
	if node.UpdatedAt >= time.Now().Add(-3*time.Hour).Unix() {
		node.Status = "up"
	} else {
		node.Status = "down"
	}
	return node
}

func farmFromDBFarm(info db.Farm) (types.Farm, error) {
	farm := types.Farm{
		Name:              info.Name,
		FarmID:            info.FarmID,
		TwinID:            info.TwinID,
		PricingPolicyID:   info.PricingPolicyID,
		CertificationType: info.Certification,
		StellarAddress:    info.StellarAddress,
		Dedicated:         info.Dedicated,
	}
	if err := json.Unmarshal([]byte(info.PublicIps), &farm.PublicIps); err != nil {
		return farm, errors.Wrap(err, "couldn't unmarshal public ips returned from db")
	}
	return farm, nil
}

func nodeWithNestedCapacityFromDBNode(info db.Node) types.NodeWithNestedCapacity {

	node := types.NodeWithNestedCapacity{
		ID:              info.ID,
		NodeID:          int(info.NodeID),
		FarmID:          int(info.FarmID),
		TwinID:          int(info.TwinID),
		Country:         info.Country,
		GridVersion:     int(info.GridVersion),
		City:            info.City,
		Uptime:          info.Uptime,
		Created:         info.Created,
		FarmingPolicyID: int(info.FarmingPolicyID),
		UpdatedAt:       info.UpdatedAt,
		Capacity: types.CapacityResult{

			Total: types.Capacity{
				CRU: uint64(info.TotalCru),
				SRU: gridtypes.Unit(info.TotalSru),
				HRU: gridtypes.Unit(info.TotalHru),
				MRU: gridtypes.Unit(info.TotalMru),
			},
			Used: types.Capacity{
				CRU: uint64(info.UsedCru),
				SRU: gridtypes.Unit(info.UsedSru),
				HRU: gridtypes.Unit(info.UsedHru),
				MRU: gridtypes.Unit(info.UsedMru),
			},
		},
		Location: types.Location{
			Country:   info.Country,
			City:      info.City,
			Longitude: info.Longitude,
			Latitude:  info.Latitude,
		},
		PublicConfig: types.PublicConfig{
			Domain: info.Domain,
			Gw4:    info.Gw4,
			Gw6:    info.Gw6,
			Ipv4:   info.Ipv4,
			Ipv6:   info.Ipv6,
		},
		CertificationType: info.Certification,
		Dedicated:         info.Dedicated,
		RentContractID:    uint(info.RentContractID),
		RentedByTwinID:    uint(info.RentedByTwinID),
		SerialNumber:      info.SerialNumber,
	}
	if node.UpdatedAt >= time.Now().Add(-3*time.Hour).Unix() {
		node.Status = "up"
	} else {
		node.Status = "down"
	}
	return node
}

func contractFromDBContract(info db.DBContract) (types.Contract, error) {
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
	contract := types.Contract{
		ContractID: info.ContractID,
		TwinID:     info.TwinID,
		State:      info.State,
		CreatedAt:  info.CreatedAt,
		Type:       info.Type,
		Details:    details,
	}
	if err := json.Unmarshal([]byte(info.ContractBillings), &contract.Billing); err != nil {
		return contract, errors.Wrap(err, "couldn't parse contract billing")
	}
	return contract, nil
}
