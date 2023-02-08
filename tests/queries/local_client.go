package main

import (
	"sort"
	"strings"

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
func (g *GridProxyClientimpl) Nodes(filter proxytypes.NodeFilter, limit proxytypes.Limit) (res []proxytypes.Node, totalCount int, err error) {
	if limit.Page == 0 {
		limit.Page = 1
	}
	if limit.Size == 0 {
		limit.Size = 50
	}
	for _, node := range g.data.nodes {
		if nodeSatisfies(&g.data, node, filter) {
			status := STATUS_DOWN
			if isUp(node.updated_at) {
				status = STATUS_UP
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
				CertificationType: node.certification,
				UpdatedAt:         int64(node.updated_at),
				Dedicated:         g.data.farms[node.farm_id].dedicated_farm,
				RentedByTwinID:    uint(g.data.nodeRentedBy[node.node_id]),
				RentContractID:    uint(g.data.nodeRentContractID[node.node_id]),
				SerialNumber:      node.serial_number,
			})
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].NodeID < res[j].NodeID
	})
	start, end := (limit.Page-1)*limit.Size, limit.Page*limit.Size
	if len(res) == 0 {
		return
	}
	if start >= uint64(len(res)) {
		start = uint64(len(res) - 1)
	}
	if end > uint64(len(res)) {
		end = uint64(len(res))
	}
	totalCount = len(res)
	res = res[start:end]
	return
}

// Farms returns farms with the given filters and pagination parameters
func (g *GridProxyClientimpl) Farms(filter proxytypes.FarmFilter, limit proxytypes.Limit) (res []proxytypes.Farm, totalCount int, err error) {
	if limit.Page == 0 {
		limit.Page = 1
	}
	if limit.Size == 0 {
		limit.Size = 50
	}
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
				CertificationType: farm.certification,
			})
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].FarmID < res[j].FarmID
	})
	start, end := (limit.Page-1)*limit.Size, limit.Page*limit.Size
	if len(res) == 0 {
		return
	}
	if start >= uint64(len(res)) {
		start = uint64(len(res) - 1)
	}
	if end > uint64(len(res)) {
		end = uint64(len(res))
	}
	totalCount = len(res)
	res = res[start:end]
	return
}

// Contracts returns contracts with the given filters and pagination parameters
func (g *GridProxyClientimpl) Contracts(filter proxytypes.ContractFilter, limit proxytypes.Limit) (res []proxytypes.Contract, totalCount int, err error) {
	if limit.Page == 0 {
		limit.Page = 1
	}
	if limit.Size == 0 {
		limit.Size = 50
	}
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
		if nodeContractsSatisfies(contract, filter) {
			contract := proxytypes.Contract{
				ContractID: uint(contract.contract_id),
				TwinID:     uint(contract.twin_id),
				State:      contract.state,
				CreatedAt:  uint(contract.created_at),
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
		if rentContractsSatisfies(contract, filter) {
			contract := proxytypes.Contract{
				ContractID: uint(contract.contract_id),
				TwinID:     uint(contract.twin_id),
				State:      contract.state,
				CreatedAt:  uint(contract.created_at),
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
		if nameContractsSatisfies(contract, filter) {
			contract := proxytypes.Contract{
				ContractID: uint(contract.contract_id),
				TwinID:     uint(contract.twin_id),
				State:      contract.state,
				CreatedAt:  uint(contract.created_at),
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
	start, end := (limit.Page-1)*limit.Size, limit.Page*limit.Size
	if len(res) == 0 {
		return
	}
	if start >= uint64(len(res)) {
		start = uint64(len(res) - 1)
	}
	if end > uint64(len(res)) {
		end = uint64(len(res))
	}
	totalCount = len(res)
	res = res[start:end]
	return
}

// Twins returns twins with the given filters and pagination parameters
func (g *GridProxyClientimpl) Twins(filter proxytypes.TwinFilter, limit proxytypes.Limit) (res []proxytypes.Twin, totalCount int, err error) {
	if limit.Page == 0 {
		limit.Page = 1
	}
	if limit.Size == 0 {
		limit.Size = 50
	}
	for _, twin := range g.data.twins {
		if twinSatisfies(twin, filter) {
			res = append(res, proxytypes.Twin{
				TwinID:    uint(twin.twin_id),
				AccountID: twin.account_id,
				Relay:     twin.relay,
				PublicKey: twin.public_key,
			})
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].TwinID < res[j].TwinID
	})
	start, end := (limit.Page-1)*limit.Size, limit.Page*limit.Size
	if len(res) == 0 {
		return
	}
	if start >= uint64(len(res)) {
		start = uint64(len(res) - 1)
	}
	if end > uint64(len(res)) {
		end = uint64(len(res))
	}
	totalCount = len(res)
	res = res[start:end]
	return
}
func (g *GridProxyClientimpl) Node(nodeID uint32) (res proxytypes.NodeWithNestedCapacity, err error) {
	node := g.data.nodes[uint64(nodeID)]
	status := STATUS_DOWN
	if isUp(node.updated_at) {
		status = STATUS_UP
	}
	res = proxytypes.NodeWithNestedCapacity{
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
		Capacity: proxytypes.CapacityResult{
			Total: proxytypes.Capacity{
				CRU: g.data.nodeTotalResources[node.node_id].cru,
				HRU: gridtypes.Unit(g.data.nodeTotalResources[node.node_id].hru),
				MRU: gridtypes.Unit(g.data.nodeTotalResources[node.node_id].mru),
				SRU: gridtypes.Unit(g.data.nodeTotalResources[node.node_id].sru),
			},
			Used: proxytypes.Capacity{
				CRU: g.data.nodeUsedResources[node.node_id].cru,
				HRU: gridtypes.Unit(g.data.nodeUsedResources[node.node_id].hru),
				MRU: gridtypes.Unit(g.data.nodeUsedResources[node.node_id].mru),
				SRU: gridtypes.Unit(g.data.nodeUsedResources[node.node_id].sru),
			},
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
		CertificationType: node.certification,
		UpdatedAt:         int64(node.updated_at),
		Dedicated:         g.data.farms[node.farm_id].dedicated_farm,
		RentedByTwinID:    uint(g.data.nodeRentedBy[node.node_id]),
		RentContractID:    uint(g.data.nodeRentContractID[node.node_id]),
		SerialNumber:      node.serial_number,
	}
	return
}

func (g *GridProxyClientimpl) NodeStatus(nodeID uint32) (res proxytypes.NodeStatus, err error) {
	node := g.data.nodes[uint64(nodeID)]
	res.Status = STATUS_DOWN
	if isUp(node.updated_at) {
		res.Status = STATUS_UP
	}
	return
}

func (g *GridProxyClientimpl) Counters(filter proxytypes.StatsFilter) (res proxytypes.Counters, err error) {
	res.Farms = int64(len(g.data.farms))
	res.Twins = int64(len(g.data.twins))
	res.PublicIPs = int64(len(g.data.publicIPs))
	res.Contracts = int64(len(g.data.rentContracts))
	res.Contracts += int64(len(g.data.nodeContracts))
	res.Contracts += int64(len(g.data.nameContracts))
	distribution := map[string]int64{}
	for _, node := range g.data.nodes {
		if filter.Status == nil || (*filter.Status == STATUS_UP && isUp(node.updated_at)) {
			res.Nodes++
			distribution[node.country] += 1
			res.TotalCRU += int64(g.data.nodeTotalResources[node.node_id].cru)
			res.TotalMRU += int64(g.data.nodeTotalResources[node.node_id].mru)
			res.TotalSRU += int64(g.data.nodeTotalResources[node.node_id].sru)
			res.TotalHRU += int64(g.data.nodeTotalResources[node.node_id].hru)
			if g.data.publicConfigs[node.node_id].ipv4 != "" || g.data.publicConfigs[node.node_id].ipv6 != "" {
				res.AccessNodes++
				if g.data.publicConfigs[node.node_id].domain != "" {
					res.Gateways++
				}
			}
		}
	}
	res.Countries = int64(len(distribution))
	res.NodesDistribution = distribution

	return
}

func nodeSatisfies(data *DBData, node node, f proxytypes.NodeFilter) bool {
	if f.Status != nil && (*f.Status == STATUS_UP) != isUp(node.updated_at) {
		return false
	}
	total := data.nodeTotalResources[node.node_id]
	used := data.nodeUsedResources[node.node_id]
	free := calcFreeResources(total, used)
	if f.FreeMRU != nil && *f.FreeMRU > free.mru {
		return false
	}
	if f.FreeHRU != nil && *f.FreeHRU > free.hru {
		return false
	}
	if f.FreeSRU != nil && *f.FreeSRU > free.sru {
		return false
	}
	if f.Country != nil && !strings.EqualFold(*f.Country, node.country) {
		return false
	}
	if f.CountryContains != nil && !stringMatch(node.country, *f.CountryContains) {
		return false
	}
	if f.TotalCRU != nil && *f.TotalCRU > total.cru {
		return false
	}
	if f.TotalHRU != nil && *f.TotalHRU > total.hru {
		return false
	}
	if f.TotalMRU != nil && *f.TotalMRU > total.mru {
		return false
	}
	if f.TotalSRU != nil && *f.TotalSRU > total.sru {
		return false
	}
	if f.NodeID != nil && *f.NodeID != node.node_id {
		return false
	}
	if f.TwinID != nil && *f.TwinID != node.twin_id {
		return false
	}
	if f.CityContains != nil && !stringMatch(node.city, *f.CityContains) {
		return false
	}
	if f.City != nil && !strings.EqualFold(*f.City, node.city) {
		return false
	}
	if f.FarmNameContains != nil && !stringMatch(data.farms[node.farm_id].name, *f.FarmNameContains) {
		return false
	}
	if f.FarmName != nil && !strings.EqualFold(*f.FarmName, data.farms[node.farm_id].name) {
		return false
	}
	if f.FarmIDs != nil && !isIn(f.FarmIDs, node.farm_id) {
		return false
	}
	if f.FreeIPs != nil && *f.FreeIPs > data.FreeIPs[node.farm_id] {
		return false
	}
	if f.IPv4 != nil && *f.IPv4 && data.publicConfigs[node.node_id].ipv4 == "" {
		return false
	}
	if f.IPv6 != nil && *f.IPv6 && data.publicConfigs[node.node_id].ipv6 == "" {
		return false
	}
	if f.Domain != nil && *f.Domain && data.publicConfigs[node.node_id].domain == "" {
		return false
	}
	rentable := data.nodeRentedBy[node.node_id] == 0 &&
		(data.farms[node.farm_id].dedicated_farm || len(data.nonDeletedContracts[node.node_id]) == 0)
	if f.Rentable != nil && *f.Rentable != rentable {
		return false
	}
	if f.RentedBy != nil && *f.RentedBy != data.nodeRentedBy[node.node_id] {
		return false
	}
	if f.AvailableFor != nil &&
		((data.nodeRentedBy[node.node_id] != 0 && data.nodeRentedBy[node.node_id] != *f.AvailableFor) ||
			(data.nodeRentedBy[node.node_id] != *f.AvailableFor && data.farms[node.farm_id].dedicated_farm)) {
		return false
	}
	if f.Rented != nil {
		_, ok := data.nodeRentedBy[node.node_id]
		return ok == *f.Rented
	}
	return true
}

func twinSatisfies(twin twin, f proxytypes.TwinFilter) bool {
	if f.TwinID != nil && twin.twin_id != *f.TwinID {
		return false
	}
	if f.AccountID != nil && twin.account_id != *f.AccountID {
		return false
	}
	return true
}

func farmSatisfies(data *DBData, farm farm, f proxytypes.FarmFilter) bool {
	if f.FreeIPs != nil && *f.FreeIPs > data.FreeIPs[farm.farm_id] {
		return false
	}
	if f.TotalIPs != nil && *f.TotalIPs > data.TotalIPs[farm.farm_id] {
		return false
	}
	if f.StellarAddress != nil && *f.StellarAddress != farm.stellar_address {
		return false
	}
	if f.PricingPolicyID != nil && *f.PricingPolicyID != farm.pricing_policy_id {
		return false
	}
	if f.FarmID != nil && *f.FarmID != farm.farm_id {
		return false
	}
	if f.TwinID != nil && *f.TwinID != farm.twin_id {
		return false
	}
	if f.NameContains != nil && *f.NameContains != "" && !stringMatch(farm.name, *f.NameContains) {
		return false
	}
	if f.Name != nil && *f.Name != "" && !strings.EqualFold(*f.Name, farm.name) {
		return false
	}
	if f.NameContains != nil && *f.NameContains != "" && !strings.Contains(farm.name, *f.NameContains) {
		return false
	}
	if f.CertificationType != nil && *f.CertificationType != "" && *f.CertificationType != farm.certification {
		return false
	}
	if f.Dedicated != nil && *f.Dedicated != farm.dedicated_farm {
		return false
	}
	return true
}

func rentContractsSatisfies(contract rent_contract, f proxytypes.ContractFilter) bool {
	if f.ContractID != nil && contract.contract_id != *f.ContractID {
		return false
	}
	if f.TwinID != nil && contract.twin_id != *f.TwinID {
		return false
	}
	if f.NodeID != nil && contract.node_id != *f.NodeID {
		return false
	}
	if f.Type != nil && *f.Type != "rent" {
		return false
	}
	if f.State != nil && contract.state != *f.State {
		return false
	}
	if f.Name != nil && *f.Name != "" {
		return false
	}
	if f.NumberOfPublicIps != nil && *f.NumberOfPublicIps != 0 {
		return false
	}
	if f.DeploymentData != nil && *f.DeploymentData != "" {
		return false
	}
	if f.DeploymentHash != nil && *f.DeploymentHash != "" {
		return false
	}
	return true
}

func nameContractsSatisfies(contract name_contract, f proxytypes.ContractFilter) bool {
	if f.ContractID != nil && contract.contract_id != *f.ContractID {
		return false
	}
	if f.TwinID != nil && contract.twin_id != *f.TwinID {
		return false
	}
	if f.NodeID != nil {
		return false
	}
	if f.Type != nil && *f.Type != "name" {
		return false
	}
	if f.State != nil && contract.state != *f.State {
		return false
	}
	if f.Name != nil && *f.Name != contract.name {
		return false
	}
	if f.NumberOfPublicIps != nil && *f.NumberOfPublicIps != 0 {
		return false
	}
	if f.DeploymentData != nil && *f.DeploymentData != "" {
		return false
	}
	if f.DeploymentHash != nil && *f.DeploymentHash != "" {
		return false
	}
	return true
}

func nodeContractsSatisfies(contract node_contract, f proxytypes.ContractFilter) bool {
	if f.ContractID != nil && contract.contract_id != *f.ContractID {
		return false
	}
	if f.TwinID != nil && contract.twin_id != *f.TwinID {
		return false
	}
	if f.NodeID != nil && contract.node_id != *f.NodeID {
		return false
	}
	if f.Type != nil && *f.Type != "node" {
		return false
	}
	if f.State != nil && contract.state != *f.State {
		return false
	}
	if f.Name != nil && *f.Name != "" {
		return false
	}
	if f.NumberOfPublicIps != nil && contract.number_of_public_i_ps < *f.NumberOfPublicIps { // TODO: fix
		return false
	}
	if f.DeploymentData != nil && contract.deployment_data != *f.DeploymentData {
		return false
	}
	if f.DeploymentHash != nil && contract.deployment_hash != *f.DeploymentHash {
		return false
	}
	return true
}
