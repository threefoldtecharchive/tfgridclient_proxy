package main

import (
	"math"
	"sort"

	proxyclient "github.com/threefoldtech/grid_proxy_server/pkg/client"
	proxytypes "github.com/threefoldtech/grid_proxy_server/pkg/types"
	"github.com/threefoldtech/zos/pkg/gridtypes"
)

// GridProxyClientimpl client that returns data directly from the db
type GridProxyClientimpl struct {
	data DBData
}

// NewGridProxyClient local grid proxy client constructor
func NewGridProxyClient(data DBData) proxyclient.Client {
	proxy := GridProxyClientimpl{data}
	return &proxy
}

// Ping makes sure the server is up
func (g *GridProxyClientimpl) Ping() error {
	return nil
}

// Nodes returns nodes with the given filters and pagination parameters
func (g *GridProxyClientimpl) Nodes(filter proxytypes.NodeFilter, limit proxytypes.Limit) (res []proxytypes.Node, err error) {
	for _, node := range g.data.nodes {
		if nodeSatisfies(&g.data, node, filter) {
			status := "down"
			if isUp(node.updated_at) {
				status = "up"
			}
			res = append(res, proxytypes.Node{
				ID:              node.id,
				NodeID:          int(node.node_id),
				FarmID:          int(node.farm_id),
				TwinID:          int(node.twin_id),
				Country:         node.country,
				City:            node.city,
				GridVersion:     int(node.grid_version),
				Uptime:          int64(node.uptime),
				Created:         int64(node.created),
				FarmingPolicyID: int(node.farming_policy_id),
				TotalResources: proxytypes.Capacity{
					CRU: g.data.nodeTotalResources[node.node_id].cru,
					HRU: gridtypes.Unit(g.data.nodeTotalResources[node.node_id].hru),
					MRU: gridtypes.Unit(g.data.nodeTotalResources[node.node_id].mru),
					SRU: gridtypes.Unit(g.data.nodeTotalResources[node.node_id].sru),
				},
				UsedResources: proxytypes.Capacity{
					CRU: g.data.nodeUsedResources[node.node_id].cru,
					HRU: gridtypes.Unit(g.data.nodeUsedResources[node.node_id].hru),
					MRU: gridtypes.Unit(g.data.nodeUsedResources[node.node_id].mru),
					SRU: gridtypes.Unit(g.data.nodeUsedResources[node.node_id].sru),
				},
				Location: proxytypes.Location{
					Country: node.country,
					City:    node.city,
				},
				PublicConfig: proxytypes.PublicConfig{
					Domain: g.data.publicConfigs[node.node_id].domain,
					Ipv4:   g.data.publicConfigs[node.node_id].ipv4,
					Ipv6:   g.data.publicConfigs[node.node_id].ipv6,
					Gw4:    g.data.publicConfigs[node.node_id].gw4,
					Gw6:    g.data.publicConfigs[node.node_id].gw6,
				},
				Status:            status,
				CertificationType: node.certification_type,
				UpdatedAt:         int64(math.Round(float64(node.updated_at) / 1000.0)),
				Dedicated:         g.data.farms[node.farm_id].dedicated_farm,
				RentedByTwinID:    uint(g.data.nodeRentedBy[node.node_id]),
				RentContractID:    uint(g.data.nodeRentContractID[node.node_id]),
			})
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].NodeID < res[j].NodeID
	})
	return
}

// Farms returns farms with the given filters and pagination parameters
func (g *GridProxyClientimpl) Farms(filter proxytypes.FarmFilter, limit proxytypes.Limit) (res []proxytypes.Farm, err error) {
	publicIPs := make(map[uint64][]proxytypes.PublicIP)
	for _, publicIP := range g.data.publicIPs {
		publicIPs[g.data.farmIDMap[publicIP.farm_id]] = append(publicIPs[g.data.farmIDMap[publicIP.farm_id]], proxytypes.PublicIP{
			ID:         publicIP.id,
			IP:         publicIP.ip,
			ContractID: int(publicIP.contract_id),
			Gateway:    publicIP.gateway,
		})
	}
	for _, farm := range g.data.farms {
		if farmSatisfies(&g.data, farm, filter) {
			res = append(res, proxytypes.Farm{
				Name:              farm.name,
				FarmID:            int(farm.farm_id),
				TwinID:            int(farm.twin_id),
				PricingPolicyID:   int(farm.pricing_policy_id),
				StellarAddress:    farm.stellar_address,
				PublicIps:         publicIPs[farm.farm_id],
				Dedicated:         farm.dedicated_farm,
				CertificationType: farm.certification_type,
			})
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].FarmID < res[j].FarmID
	})
	return

}

// Contracts returns contracts with the given filters and pagination parameters
func (g *GridProxyClientimpl) Contracts(filter proxytypes.ContractFilter, limit proxytypes.Limit) (res []proxytypes.Contract, err error) {
	billings := make(map[uint64][]proxytypes.ContractBilling)
	for contractID, contractBillings := range g.data.billings {
		for _, billing := range contractBillings {
			billings[contractID] = append(billings[contractID], proxytypes.ContractBilling{
				AmountBilled:     billing.amount_billed,
				DiscountReceived: billing.discount_received,
				Timestamp:        billing.timestamp,
			})
		}
		sort.Slice(billings[contractID], func(i, j int) bool {
			return billings[contractID][i].Timestamp < billings[contractID][j].Timestamp
		})
	}
	for _, contract := range g.data.nodeContracts {
		if nodeContractsSatisfies(&g.data, contract, filter) {
			contract := proxytypes.Contract{
				ContractID: uint(contract.contract_id),
				TwinID:     uint(contract.twin_id),
				State:      contract.state,
				CreatedAt:  uint(math.Round(float64(contract.created_at) / 1000.0)),
				Type:       "node",
				Details: proxytypes.NodeContractDetails{
					NodeID:            uint(contract.node_id),
					DeploymentData:    contract.deployment_data,
					DeploymentHash:    contract.deployment_hash,
					NumberOfPublicIps: uint(contract.number_of_public_i_ps),
				},
				Billing: billings[contract.contract_id],
			}
			res = append(res, contract)
		}
	}
	for _, contract := range g.data.rentContracts {
		if rentContractsSatisfies(&g.data, contract, filter) {
			contract := proxytypes.Contract{
				ContractID: uint(contract.contract_id),
				TwinID:     uint(contract.twin_id),
				State:      contract.state,
				CreatedAt:  uint(math.Round(float64(contract.created_at) / 1000.0)),
				Type:       "rent",
				Details: proxytypes.RentContractDetails{
					NodeID: uint(contract.node_id),
				},
				Billing: billings[contract.contract_id],
			}
			res = append(res, contract)
		}
	}
	for _, contract := range g.data.nameContracts {
		if nameContractsSatisfies(&g.data, contract, filter) {
			contract := proxytypes.Contract{
				ContractID: uint(contract.contract_id),
				TwinID:     uint(contract.twin_id),
				State:      contract.state,
				CreatedAt:  uint(math.Round(float64(contract.created_at) / 1000.0)),
				Type:       "name",
				Details: proxytypes.NameContractDetails{
					Name: contract.name,
				},
				Billing: billings[contract.contract_id],
			}
			res = append(res, contract)
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].ContractID < res[j].ContractID
	})
	return

}

func (g *GridProxyClientimpl) Node(nodeID uint32) (res proxytypes.NodeWithNestedCapacity, err error) {
	return
}

func (g *GridProxyClientimpl) NodeStatus(nodeID uint32) (res proxytypes.NodeStatus, err error) {

	return
}
