package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/threefoldtech/zos/pkg/gridtypes"
)

var (
	nodesMRU               = make(map[uint64]uint64)
	nodesSRU               = make(map[uint64]uint64)
	nodesHRU               = make(map[uint64]uint64)
	nodeUP                 = make(map[uint64]bool)
	createdNodeContracts   = make([]uint64, 0)
	dedicatedFarms         = make(map[uint64]struct{})
	availableRentNodes     = make(map[uint64]struct{})
	availableRentNodesList = make([]uint64, 0)
	renter                 = make(map[uint64]uint64)
	billCnt                = 1
	contractCnt            = uint64(1)
)

const (
	contractCreatedRatio = .1 // from devnet
	usedPublicIPsRatio   = .9
	nodeUpRatio          = .5
	nodeCount            = 1000
	farmCount            = 100
	normalUsers          = 2000
	publicIPCount        = 1000
	twinCount            = nodeCount + farmCount + normalUsers
	contractCount        = 3000
	rentContractCount    = 100
	nameContractCount    = 300

	maxContractHRU = 1024 * 1024 * 1024 * 300
	maxContractSRU = 1024 * 1024 * 1024 * 300
	maxContractMRU = 1024 * 1024 * 1024 * 16
	maxContractCRU = 16
	minContractHRU = 0
	minContractSRU = 1024 * 1024 * 256
	minContractMRU = 1024 * 1024 * 256
	minContractCRU = 1
)

func initSchema(db *sql.DB) error {
	schema, err := os.ReadFile("./schema.sql")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(string(schema))
	if err != nil {
		panic(err)
	}
	return nil
}

func generateTwins(db *sql.DB) error {
	for i := uint64(1); i <= twinCount; i++ {
		twin := twin{
			id:           fmt.Sprintf("twin-%d", i),
			account_id:   fmt.Sprintf("account-id-%d", i),
			relay:        fmt.Sprintf("relay-%d", i),
			public_key:   fmt.Sprintf("public-key-%d", i),
			twin_id:      i,
			grid_version: 3,
		}
		if _, err := db.Exec(insertQuery(&twin)); err != nil {
			panic(err)
		}
	}
	return nil
}

func generatePublicIPs(db *sql.DB) error {
	for i := uint64(1); i <= publicIPCount; i++ {
		contract_id := uint64(0)
		if flip(usedPublicIPsRatio) {
			contract_id = createdNodeContracts[rnd(0, uint64(len(createdNodeContracts))-1)]
		}
		ip := randomIPv4()
		public_ip := public_ip{
			id:          fmt.Sprintf("public-ip-%d", i),
			gateway:     ip.String(),
			ip:          IPv4Subnet(ip).String(),
			contract_id: contract_id,
			farm_id:     fmt.Sprintf("farm-%d", rnd(1, farmCount)),
		}
		if _, err := db.Exec(insertQuery(&public_ip)); err != nil {
			panic(err)
		}
		if _, err := db.Exec(fmt.Sprintf("UPDATE node_contract set number_of_public_i_ps = number_of_public_i_ps + 1 WHERE contract_id = %d;", contract_id)); err != nil {
			panic(err)
		}
	}
	return nil
}

func generateFarms(db *sql.DB) error {
	for i := uint64(1); i <= farmCount; i++ {
		farm := farm{
			id:                fmt.Sprintf("farm-%d", i),
			farm_id:           i,
			name:              fmt.Sprintf("farm-name-%d", i),
			certification:     "Diy",
			dedicated_farm:    flip(.1),
			twin_id:           i,
			pricing_policy_id: 1,
			grid_version:      3,
			stellar_address:   "",
		}
		if farm.dedicated_farm {
			dedicatedFarms[farm.farm_id] = struct{}{}
		}
		if _, err := db.Exec(insertQuery(&farm)); err != nil {
			panic(err)
		}
	}
	return nil
}

func generateContracts(db *sql.DB) error {
	for i := uint64(1); i <= contractCount; i++ {
		nodeID := rnd(1, nodeCount)
		state := "Deleted"
		if nodeUP[nodeID] {
			if flip(contractCreatedRatio) {
				state = "Created"
			} else if flip(0.5) {
				state = "GracePeriod"
			}
		}
		if state != "Deleted" && (minContractHRU > nodesHRU[nodeID] || minContractMRU > nodesMRU[nodeID] || minContractSRU > nodesSRU[nodeID]) {
			i--
			continue
		}
		twinID := rnd(1100, 3100)
		if renter, ok := renter[nodeID]; ok {
			twinID = renter
		}
		if _, ok := availableRentNodes[nodeID]; ok {
			i--
			continue
		}
		contract := node_contract{
			id:                    fmt.Sprintf("node-contract-%d", contractCnt),
			twin_id:               twinID,
			contract_id:           contractCnt,
			state:                 state,
			created_at:            uint64(time.Now().Unix()),
			node_id:               nodeID,
			deployment_data:       fmt.Sprintf("deployment-data-%d", contractCnt),
			deployment_hash:       fmt.Sprintf("deployment-hash-%d", contractCnt),
			number_of_public_i_ps: 0,
			grid_version:          3,
			resources_used_id:     "",
		}
		cru := rnd(minContractCRU, maxContractCRU)
		hru := rnd(minContractHRU, min(maxContractHRU, nodesHRU[nodeID]))
		sru := rnd(minContractSRU, min(maxContractSRU, nodesSRU[nodeID]))
		mru := rnd(minContractMRU, min(maxContractMRU, nodesMRU[nodeID]))
		contract_resources := contract_resources{
			id:          fmt.Sprintf("contract-resources-%d", contractCnt),
			hru:         hru,
			sru:         sru,
			cru:         cru,
			mru:         mru,
			contract_id: fmt.Sprintf("node-contract-%d", contractCnt),
		}
		if contract.state != "Deleted" {
			nodesHRU[nodeID] -= hru
			nodesSRU[nodeID] -= sru
			nodesMRU[nodeID] -= mru
			createdNodeContracts = append(createdNodeContracts, contractCnt)
		}
		if _, err := db.Exec(insertQuery(&contract)); err != nil {
			panic(err)
		}
		if _, err := db.Exec(insertQuery(&contract_resources)); err != nil {
			panic(err)
		}
		if _, err := db.Exec(fmt.Sprintf(`UPDATE node_contract SET resources_used_id = 'contract-resources-%d' WHERE id = 'node-contract-%d'`, contractCnt, contractCnt)); err != nil {
			panic(err)
		}
		billings := rnd(0, 10)
		for j := uint64(0); j < billings; j++ {
			billing := contract_bill_report{
				id:                fmt.Sprintf("contract-bill-report-%d", billCnt),
				contract_id:       contractCnt,
				discount_received: "Default",
				amount_billed:     rnd(0, 100000),
				timestamp:         uint64(time.Now().UnixNano()),
			}
			billCnt++
			if _, err := db.Exec(insertQuery(&billing)); err != nil {
				panic(err)
			}
		}
		contractCnt++
	}
	return nil
}
func generateNameContracts(db *sql.DB) error {
	for i := uint64(1); i <= nameContractCount; i++ {
		nodeID := rnd(1, nodeCount)
		state := "Deleted"
		if nodeUP[nodeID] {
			if flip(contractCreatedRatio) {
				state = "Created"
			} else if flip(0.5) {
				state = "GracePeriod"
			}
		}
		twinID := rnd(1100, 3100)
		if renter, ok := renter[nodeID]; ok {
			twinID = renter
		}
		if _, ok := availableRentNodes[nodeID]; ok {
			i--
			continue
		}
		contract := name_contract{
			id:           fmt.Sprintf("node-contract-%d", contractCnt),
			twin_id:      twinID,
			contract_id:  contractCnt,
			state:        state,
			created_at:   uint64(time.Now().Unix()),
			grid_version: 3,
			name:         uuid.NewString(),
		}
		if _, err := db.Exec(insertQuery(&contract)); err != nil {
			panic(err)
		}
		billings := rnd(0, 10)
		for j := uint64(0); j < billings; j++ {
			billing := contract_bill_report{
				id:                fmt.Sprintf("contract-bill-report-%d", billCnt),
				contract_id:       contractCnt,
				discount_received: "Default",
				amount_billed:     rnd(0, 100000),
				timestamp:         uint64(time.Now().UnixNano()),
			}
			billCnt++
			if _, err := db.Exec(insertQuery(&billing)); err != nil {
				panic(err)
			}
		}
		contractCnt++
	}
	return nil
}
func generateRentContracts(db *sql.DB) error {
	for i := uint64(1); i <= rentContractCount; i++ {
		nl, nodeID := popRandom(availableRentNodesList)
		availableRentNodesList = nl
		delete(availableRentNodes, nodeID)
		state := "Deleted"
		if nodeUP[nodeID] {
			if flip(0.9) {
				state = "Created"
			} else if flip(0.5) {
				state = "GracePeriod"
			}
		}
		contract := rent_contract{
			id:           fmt.Sprintf("rent-contract-%d", contractCnt),
			twin_id:      rnd(1100, 3100),
			contract_id:  contractCnt,
			state:        state,
			created_at:   uint64(time.Now().Unix()),
			node_id:      nodeID,
			grid_version: 3,
		}
		if state != "Deleted" {
			renter[nodeID] = contract.twin_id
		}
		if _, err := db.Exec(insertQuery(&contract)); err != nil {
			panic(err)
		}
		billings := rnd(0, 10)
		for j := uint64(0); j < billings; j++ {
			billing := contract_bill_report{
				id:                fmt.Sprintf("contract-bill-report-%d", billCnt),
				contract_id:       contractCnt,
				discount_received: "Default",
				amount_billed:     rnd(0, 100000),
				timestamp:         uint64(time.Now().UnixNano()),
			}

			billCnt++
			if _, err := db.Exec(insertQuery(&billing)); err != nil {
				panic(err)
			}
		}
		contractCnt++
	}
	return nil
}

func generateNodes(db *sql.DB) error {
	const NodeCount = 1000
	for i := uint64(1); i <= NodeCount; i++ {
		mru := rnd(4, 256) * 1024 * 1024 * 1024
		hru := rnd(100, 30*1024) * 1024 * 1024 * 1024 // 100GB -> 30TB
		sru := rnd(200, 30*1024) * 1024 * 1024 * 1024 // 100GB -> 30TB
		cru := rnd(4, 128)
		up := flip(nodeUpRatio)
		updatedAt := time.Now().Unix() - int64(rnd(60*60*3, 60*60*24*30*12))
		if up {
			updatedAt = time.Now().Unix() - int64(rnd(0, 60*60*1))
		}
		nodesMRU[i] = mru - max(2*uint64(gridtypes.Gigabyte), mru/10)
		nodesSRU[i] = sru - 100*uint64(gridtypes.Gigabyte)
		nodesHRU[i] = hru
		nodeUP[i] = up
		location := location{
			id:        fmt.Sprintf("location-%d", i),
			longitude: fmt.Sprintf("location--long-%d", i),
			latitude:  fmt.Sprintf("location-lat-%d", i),
		}
		node := node{
			id:                fmt.Sprintf("node-%d", i),
			location_id:       fmt.Sprintf("location-%d", i),
			node_id:           i,
			farm_id:           i%100 + 1,
			twin_id:           i + 100 + 1,
			country:           "Belgium",
			city:              "Unknown",
			uptime:            1000,
			updated_at:        uint64(updatedAt),
			created:           uint64(time.Now().Unix()),
			created_at:        uint64(time.Now().Unix()),
			farming_policy_id: 1,
			grid_version:      3,
			certification:     "Diy",
			secure:            false,
			virtualized:       false,
			serial_number:     "",
		}
		total_resources := node_resources_total{
			id:      fmt.Sprintf("total-resources-%d", i),
			hru:     hru,
			sru:     sru,
			cru:     cru,
			mru:     mru,
			node_id: fmt.Sprintf("node-%d", i),
		}
		if _, ok := dedicatedFarms[node.farm_id]; ok {
			availableRentNodes[i] = struct{}{}
			availableRentNodesList = append(availableRentNodesList, i)
		}
		if _, err := db.Exec(insertQuery(&location)); err != nil {
			panic(err)
		}
		if _, err := db.Exec(insertQuery(&node)); err != nil {
			panic(err)
		}
		if _, err := db.Exec(insertQuery(&total_resources)); err != nil {
			panic(err)
		}

		if flip(.1) {
			if _, err := db.Exec(insertQuery(&public_config{
				id:      fmt.Sprintf("public-config-%d", i),
				ipv4:    "185.16.5.2/24",
				gw4:     "185.16.5.2",
				ipv6:    "::1/64",
				gw6:     "::1",
				domain:  "hamada.com",
				node_id: fmt.Sprintf("node-%d", i),
			})); err != nil {
				panic(err)
			}
		}
	}
	return nil
}

func generateData(db *sql.DB) error {
	if err := generateTwins(db); err != nil {
		panic(err)
	}
	if err := generateFarms(db); err != nil {
		panic(err)
	}
	if err := generateNodes(db); err != nil {
		panic(err)
	}
	if err := generateRentContracts(db); err != nil {
		panic(err)
	}
	if err := generateContracts(db); err != nil {
		panic(err)
	}
	if err := generateNameContracts(db); err != nil {
		panic(err)
	}
	if err := generatePublicIPs(db); err != nil {
		panic(err)
	}
	return nil
}
