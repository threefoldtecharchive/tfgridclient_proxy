package main

import (
	"database/sql"
	"math"

	"github.com/threefoldtech/zos/pkg/gridtypes"
)

type DBData struct {
	nodeIDMap          map[string]uint64
	farmIDMap          map[string]uint64
	FreeIPs            map[uint64]uint64
	TotalIPs           map[uint64]uint64
	nodeUsedResources  map[uint64]node_resources_total
	nodeRentedBy       map[uint64]uint64
	nodeRentContractID map[uint64]uint64

	nodes               map[uint64]node
	nodeTotalResources  map[uint64]node_resources_total
	farms               map[uint64]farm
	twins               map[uint64]twin
	publicIPs           map[string]public_ip
	publicConfigs       map[uint64]public_config
	nodeContracts       map[uint64]node_contract
	rentContracts       map[uint64]rent_contract
	nameContracts       map[uint64]name_contract
	billings            map[uint64][]contract_bill_report
	contractResources   map[string]contract_resources
	nonDeletedContracts map[uint64][]uint64
	db                  *sql.DB
}

func loadNodes(db *sql.DB, data *DBData) error {
	rows, err := db.Query(`
	SELECT
		COALESCE(id, ''),
		COALESCE(grid_version, 0),
		COALESCE(node_id, 0),
		COALESCE(farm_id, 0),
		COALESCE(twin_id, 0),
		COALESCE(country, ''),
		COALESCE(city, ''),
		COALESCE(uptime, 0),
		COALESCE(created, 0),
		COALESCE(farming_policy_id, 0),
		COALESCE(certification, ''),
		COALESCE(secure, false),
		COALESCE(virtualized, false),
		COALESCE(serial_number, ''),
		COALESCE(created_at, 0),
		COALESCE(updated_at, 0),
		COALESCE(location_id, '')
	FROM
		node;`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var node node
		if err := rows.Scan(
			&node.id,
			&node.grid_version,
			&node.node_id,
			&node.farm_id,
			&node.twin_id,
			&node.country,
			&node.city,
			&node.uptime,
			&node.created,
			&node.farming_policy_id,
			&node.certification,
			&node.secure,
			&node.virtualized,
			&node.serial_number,
			&node.created_at,
			&node.updated_at,
			&node.location_id,
		); err != nil {
			return err
		}
		data.nodes[node.node_id] = node
		data.nodeIDMap[node.id] = node.node_id
	}
	return nil
}

func calcNodesUsedResources(data *DBData) error {

	for _, node := range data.nodes {
		used := node_resources_total{
			mru: uint64(2 * gridtypes.Gigabyte),
			sru: uint64(100 * gridtypes.Gigabyte),
		}
		tenpercent := uint64(math.Round(float64(data.nodeTotalResources[node.node_id].mru) / 10))
		if used.mru < tenpercent {
			used.mru = tenpercent
		}
		data.nodeUsedResources[node.node_id] = used
	}

	for _, contract := range data.nodeContracts {
		if contract.state == "Deleted" {
			continue
		}
		contratResourceID := contract.resources_used_id
		data.nodeUsedResources[contract.node_id] = node_resources_total{
			cru: data.contractResources[contratResourceID].cru + data.nodeUsedResources[contract.node_id].cru,
			mru: data.contractResources[contratResourceID].mru + data.nodeUsedResources[contract.node_id].mru,
			hru: data.contractResources[contratResourceID].hru + data.nodeUsedResources[contract.node_id].hru,
			sru: data.contractResources[contratResourceID].sru + data.nodeUsedResources[contract.node_id].sru,
		}

	}
	return nil
}

func calcRentInfo(data *DBData) error {
	for _, contract := range data.rentContracts {
		if contract.state == "Deleted" {
			continue
		}
		data.nodeRentedBy[contract.node_id] = contract.twin_id
		data.nodeRentContractID[contract.node_id] = contract.contract_id
	}
	return nil
}

func calcFreeIPs(data *DBData) error {
	for _, publicIP := range data.publicIPs {
		if publicIP.contract_id == 0 {
			data.FreeIPs[data.farmIDMap[publicIP.farm_id]]++
		}
		data.TotalIPs[data.farmIDMap[publicIP.farm_id]]++
	}
	return nil
}

func loadNodesTotalResources(db *sql.DB, data *DBData) error {
	rows, err := db.Query(`
	SELECT
		COALESCE(id, ''),
		COALESCE(hru, 0),
		COALESCE(sru, 0),
		COALESCE(cru, 0),
		COALESCE(mru, 0),
		COALESCE(node_id, '')
	FROM
		node_resources_total;`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var nodeResourcesTotal node_resources_total
		if err := rows.Scan(
			&nodeResourcesTotal.id,
			&nodeResourcesTotal.hru,
			&nodeResourcesTotal.sru,
			&nodeResourcesTotal.cru,
			&nodeResourcesTotal.mru,
			&nodeResourcesTotal.node_id,
		); err != nil {
			return err
		}
		data.nodeTotalResources[data.nodeIDMap[nodeResourcesTotal.node_id]] = nodeResourcesTotal
	}
	return nil
}

func loadFarms(db *sql.DB, data *DBData) error {
	rows, err := db.Query(`
	SELECT 
		COALESCE(id, ''),
		COALESCE(grid_version, 0),
		COALESCE(farm_id, 0),
		COALESCE(name, ''),
		COALESCE(twin_id, 0),
		COALESCE(pricing_policy_id, 0),
		COALESCE(certification, ''),
		COALESCE(stellar_address, ''),
		COALESCE(dedicated_farm, false)
	FROM
		farm;`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var farm farm
		if err := rows.Scan(
			&farm.id,
			&farm.grid_version,
			&farm.farm_id,
			&farm.name,
			&farm.twin_id,
			&farm.pricing_policy_id,
			&farm.certification,
			&farm.stellar_address,
			&farm.dedicated_farm,
		); err != nil {
			return err
		}
		data.farms[farm.farm_id] = farm
		data.farmIDMap[farm.id] = farm.farm_id
	}
	return nil
}

func loadTwins(db *sql.DB, data *DBData) error {
	rows, err := db.Query(`
	SELECT
	COALESCE(id, ''),
	COALESCE(grid_version, 0),
	COALESCE(twin_id, 0),
	COALESCE(account_id, ''),
	COALESCE(relay, ''),
	COALESCE(public_key, '')
	FROM
		twin;`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var twin twin
		if err := rows.Scan(
			&twin.id,
			&twin.grid_version,
			&twin.twin_id,
			&twin.account_id,
			&twin.relay,
			&twin.public_key,
		); err != nil {
			return err
		}
		data.twins[twin.twin_id] = twin
	}
	return nil
}

func loadPublicIPs(db *sql.DB, data *DBData) error {
	rows, err := db.Query(`
	SELECT 
		COALESCE(id, ''),
		COALESCE(gateway, ''),
		COALESCE(ip, ''),
		COALESCE(contract_id, 0),
		COALESCE(farm_id, '')
	FROM
		public_ip;`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var publicIP public_ip
		if err := rows.Scan(
			&publicIP.id,
			&publicIP.gateway,
			&publicIP.ip,
			&publicIP.contract_id,
			&publicIP.farm_id,
		); err != nil {
			return err
		}
		data.publicIPs[publicIP.id] = publicIP
	}
	return nil
}

func loadPublicConfigs(db *sql.DB, data *DBData) error {
	rows, err := db.Query(`
	SELECT
	COALESCE(id, ''),
	COALESCE(ipv4, ''),
	COALESCE(ipv6, ''),
	COALESCE(gw4, ''),
	COALESCE(gw6, ''),
	COALESCE(domain, ''),
	COALESCE(node_id, '')
	FROM
		public_config;`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var publicConfig public_config
		if err := rows.Scan(
			&publicConfig.id,
			&publicConfig.ipv4,
			&publicConfig.ipv6,
			&publicConfig.gw4,
			&publicConfig.gw6,
			&publicConfig.domain,
			&publicConfig.node_id,
		); err != nil {
			return err
		}
		data.publicConfigs[data.nodeIDMap[publicConfig.node_id]] = publicConfig
	}
	return nil
}
func loadContracts(db *sql.DB, data *DBData) error {
	rows, err := db.Query(`
	SELECT
		COALESCE(id, ''),
		COALESCE(grid_version, 0),
		COALESCE(contract_id, 0),
		COALESCE(twin_id, 0),
		COALESCE(node_id, 0),
		COALESCE(deployment_data, ''),
		COALESCE(deployment_hash, ''),
		COALESCE(number_of_public_i_ps, 0),
		COALESCE(state, ''),
		COALESCE(created_at, 0),
		COALESCE(resources_used_id, '')
	FROM
		node_contract;`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var contract node_contract
		if err := rows.Scan(
			&contract.id,
			&contract.grid_version,
			&contract.contract_id,
			&contract.twin_id,
			&contract.node_id,
			&contract.deployment_data,
			&contract.deployment_hash,
			&contract.number_of_public_i_ps,
			&contract.state,
			&contract.created_at,
			&contract.resources_used_id,
		); err != nil {
			return err
		}
		data.nodeContracts[contract.contract_id] = contract
		if contract.state != "Deleted" {
			data.nonDeletedContracts[contract.node_id] = append(data.nonDeletedContracts[contract.node_id], contract.contract_id)
		}

	}
	return nil
}
func loadRentContracts(db *sql.DB, data *DBData) error {
	rows, err := db.Query(`
	SELECT
		COALESCE(id, ''),
		COALESCE(grid_version, 0),
		COALESCE(contract_id, 0),
		COALESCE(twin_id, 0),
		COALESCE(node_id, 0),
		COALESCE(state, ''),
		COALESCE(created_at, 0)
	FROM
		rent_contract;`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var contract rent_contract
		if err := rows.Scan(

			&contract.id,
			&contract.grid_version,
			&contract.contract_id,
			&contract.twin_id,
			&contract.node_id,
			&contract.state,
			&contract.created_at,
		); err != nil {
			return err
		}
		data.rentContracts[contract.contract_id] = contract
	}
	return nil
}
func loadNameContracts(db *sql.DB, data *DBData) error {
	rows, err := db.Query(`
	SELECT
		COALESCE(id, ''),
		COALESCE(grid_version, 0),
		COALESCE(contract_id, 0),
		COALESCE(twin_id, 0),
		COALESCE(name, ''),
		COALESCE(state, ''),
		COALESCE(created_at, 0)
	FROM
		name_contract;`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var contract name_contract
		if err := rows.Scan(
			&contract.id,
			&contract.grid_version,
			&contract.contract_id,
			&contract.twin_id,
			&contract.name,
			&contract.state,
			&contract.created_at,
		); err != nil {
			return err
		}
		data.nameContracts[contract.contract_id] = contract
	}
	return nil
}

func loadContractResources(db *sql.DB, data *DBData) error {
	rows, err := db.Query(`
	SELECT
	COALESCE(id, ''),
	COALESCE(hru, 0),
	COALESCE(sru, 0),
	COALESCE(cru, 0),
	COALESCE(mru, 0),
	COALESCE(contract_id, '')
	FROM
		contract_resources;`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var contractResources contract_resources
		if err := rows.Scan(
			&contractResources.id,
			&contractResources.hru,
			&contractResources.sru,
			&contractResources.cru,
			&contractResources.mru,
			&contractResources.contract_id,
		); err != nil {
			return err
		}
		data.contractResources[contractResources.id] = contractResources
	}
	return nil
}
func loadContractBillingReports(db *sql.DB, data *DBData) error {
	rows, err := db.Query(`
	SELECT
		COALESCE(id, ''),
		COALESCE(contract_id, 0),
		COALESCE(discount_received, ''),
		COALESCE(amount_billed, 0),
		COALESCE(timestamp, 0)
	FROM
		contract_bill_report;`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var contractBillReport contract_bill_report
		if err := rows.Scan(
			&contractBillReport.id,
			&contractBillReport.contract_id,
			&contractBillReport.discount_received,
			&contractBillReport.amount_billed,
			&contractBillReport.timestamp,
		); err != nil {
			return err
		}
		data.billings[contractBillReport.contract_id] = append(data.billings[contractBillReport.contract_id], contractBillReport)
	}
	return nil
}

func load(db *sql.DB) (DBData, error) {
	data := DBData{
		nodeIDMap:           make(map[string]uint64),
		farmIDMap:           make(map[string]uint64),
		FreeIPs:             make(map[uint64]uint64),
		TotalIPs:            make(map[uint64]uint64),
		nodes:               make(map[uint64]node),
		farms:               make(map[uint64]farm),
		twins:               make(map[uint64]twin),
		publicIPs:           make(map[string]public_ip),
		publicConfigs:       make(map[uint64]public_config),
		nodeContracts:       make(map[uint64]node_contract),
		rentContracts:       make(map[uint64]rent_contract),
		nameContracts:       make(map[uint64]name_contract),
		nodeRentedBy:        make(map[uint64]uint64),
		nodeRentContractID:  make(map[uint64]uint64),
		billings:            make(map[uint64][]contract_bill_report),
		contractResources:   make(map[string]contract_resources),
		nodeTotalResources:  make(map[uint64]node_resources_total),
		nodeUsedResources:   make(map[uint64]node_resources_total),
		nonDeletedContracts: make(map[uint64][]uint64),
		db:                  db,
	}
	if err := loadNodes(db, &data); err != nil {
		return data, err
	}
	if err := loadFarms(db, &data); err != nil {
		return data, err
	}
	if err := loadTwins(db, &data); err != nil {
		return data, err
	}
	if err := loadPublicConfigs(db, &data); err != nil {
		return data, err
	}
	if err := loadPublicIPs(db, &data); err != nil {
		return data, err
	}
	if err := loadContracts(db, &data); err != nil {
		return data, err
	}
	if err := loadRentContracts(db, &data); err != nil {
		return data, err
	}
	if err := loadNameContracts(db, &data); err != nil {
		return data, err
	}
	if err := loadContractResources(db, &data); err != nil {
		return data, err
	}
	if err := loadContractBillingReports(db, &data); err != nil {
		return data, err
	}
	if err := loadNodesTotalResources(db, &data); err != nil {
		return data, err
	}
	if err := calcNodesUsedResources(&data); err != nil {
		return data, err
	}
	if err := calcRentInfo(&data); err != nil {
		return data, err
	}
	if err := calcFreeIPs(&data); err != nil {
		return data, err
	}
	return data, nil
}
