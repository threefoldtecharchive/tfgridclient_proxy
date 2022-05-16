package main

import (
	"math"
	"sort"

	"github.com/threefoldtech/grid_proxy_server/pkg/gridproxy"
	"github.com/threefoldtech/zos/pkg/gridtypes"
)

type GridProxyClientimpl struct {
	data DBData
}

func NewGridProxyClient(data DBData) gridproxy.GridProxyClient {
	proxy := GridProxyClientimpl{data}
	return &proxy
}

func (g *GridProxyClientimpl) Ping() error {
	return nil
}

func (g *GridProxyClientimpl) Nodes(filter gridproxy.NodeFilter, limit gridproxy.Limit) (res []gridproxy.Node, err error) {
	for _, node := range g.data.nodes {
		if nodeSatisfies(&g.data, node, filter) {
			status := "down"
			if isUp(node.updated_at) {
				status = "up"
			}
			res = append(res, gridproxy.Node{
				ID:              node.id,
				NodeID:          uint32(node.node_id),
				FarmID:          int(node.farm_id),
				TwinID:          int(node.twin_id),
				Country:         node.country,
				City:            node.city,
				GridVersion:     int(node.grid_version),
				Uptime:          int64(node.uptime),
				Created:         int64(node.created),
				FarmingPolicyID: int(node.farming_policy_id),
				TotalResources: gridproxy.Capacity{
					CRU: g.data.nodeTotalResources[node.node_id].cru,
					HRU: gridtypes.Unit(g.data.nodeTotalResources[node.node_id].hru),
					MRU: gridtypes.Unit(g.data.nodeTotalResources[node.node_id].mru),
					SRU: gridtypes.Unit(g.data.nodeTotalResources[node.node_id].sru),
				},
				UsedResources: gridproxy.Capacity{
					CRU: g.data.nodeUsedResources[node.node_id].cru,
					HRU: gridtypes.Unit(g.data.nodeUsedResources[node.node_id].hru),
					MRU: gridtypes.Unit(g.data.nodeUsedResources[node.node_id].mru),
					SRU: gridtypes.Unit(g.data.nodeUsedResources[node.node_id].sru),
				},
				Location: gridproxy.Location{
					Country: node.country,
					City:    node.city,
				},
				PublicConfig: gridproxy.PublicConfig{
					Domain: g.data.publicConfigs[node.node_id].domain,
					Ipv4:   g.data.publicConfigs[node.node_id].ipv4,
					Ipv6:   g.data.publicConfigs[node.node_id].ipv6,
					Gw4:    g.data.publicConfigs[node.node_id].gw4,
					Gw6:    g.data.publicConfigs[node.node_id].gw6,
				},
				Status:            status,
				CertificationType: node.certification_type,
			})
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].NodeID < res[j].NodeID
	})
	return
}

func (g *GridProxyClientimpl) Farms(filter gridproxy.FarmFilter, limit gridproxy.Limit) (res gridproxy.FarmResult, err error) {
	publicIPs := make(map[uint64][]gridproxy.PublicIP)
	for _, publicIP := range g.data.publicIPs {
		publicIPs[g.data.farmIDMap[publicIP.farm_id]] = append(publicIPs[g.data.farmIDMap[publicIP.farm_id]], gridproxy.PublicIP{
			ID:         publicIP.id,
			IP:         publicIP.ip,
			ContractID: int(publicIP.contract_id),
			Gateway:    publicIP.gateway,
		})
	}
	for _, farm := range g.data.farms {
		if farmSatisfies(&g.data, farm, filter) {
			res = append(res, gridproxy.Farm{
				Name:            farm.name,
				FarmID:          int(farm.farm_id),
				TwinID:          int(farm.twin_id),
				PricingPolicyID: int(farm.pricing_policy_id),
				Version:         int(farm.grid_version),
				StellarAddress:  farm.stellar_address,
				PublicIps:       publicIPs[farm.farm_id],
			})
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].FarmID < res[j].FarmID
	})
	return

}
func (g *GridProxyClientimpl) Contracts(filter gridproxy.ContractFilter, limit gridproxy.Limit) (res []gridproxy.Contract, err error) {
	billings := make(map[uint64][]gridproxy.ContractBilling)
	for contractID, contractBillings := range g.data.billings {
		for _, billing := range contractBillings {
			billings[contractID] = append(billings[contractID], gridproxy.ContractBilling{
				AmountBilled:     billing.amount_billed,
				DiscountReceived: billing.discount_received,
				Timestamp:        billing.timestamp,
			})
		}
		sort.Slice(billings[contractID], func(i, j int) bool {
			return billings[contractID][i].Timestamp < billings[contractID][j].Timestamp
		})
	}
	for _, contract := range g.data.node_contracts {
		if nodeContractsSatisfies(&g.data, contract, filter) {
			contract := gridproxy.Contract{
				ContractID: uint(contract.contract_id),
				TwinID:     uint(contract.twin_id),
				State:      contract.state,
				CreatedAt:  uint(math.Round(float64(contract.created_at) / 1000.0)),
				Type:       "node",
				Details: gridproxy.NodeContractDetails{
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
	for _, contract := range g.data.rent_contracts {
		if rentContractsSatisfies(&g.data, contract, filter) {
			contract := gridproxy.Contract{
				ContractID: uint(contract.contract_id),
				TwinID:     uint(contract.twin_id),
				State:      contract.state,
				CreatedAt:  uint(math.Round(float64(contract.created_at) / 1000.0)),
				Type:       "rent",
				Details: gridproxy.RentContractDetails{
					NodeID: uint(contract.node_id),
				},
				Billing: billings[contract.contract_id],
			}
			res = append(res, contract)
		}
	}
	for _, contract := range g.data.name_contracts {
		if nameContractsSatisfies(&g.data, contract, filter) {
			contract := gridproxy.Contract{
				ContractID: uint(contract.contract_id),
				TwinID:     uint(contract.twin_id),
				State:      contract.state,
				CreatedAt:  uint(math.Round(float64(contract.created_at) / 1000.0)),
				Type:       "name",
				Details: gridproxy.NameContractDetails{
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

func (g *GridProxyClientimpl) Node(nodeID uint32) (res gridproxy.NodeInfo, err error) {
	return
}

func (g *GridProxyClientimpl) NodeStatus(nodeID uint32) (res gridproxy.NodeStatus, err error) {

	return
}
