package main

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"

	"github.com/pkg/errors"
	proxyclient "github.com/threefoldtech/grid_proxy_server/pkg/client"
	proxytypes "github.com/threefoldtech/grid_proxy_server/pkg/types"
)

var (
	contractsReturned = make(map[int]uint64)
)

const (
	ContractsTests = 2000
)

type ContractsAggregate struct {
	contractIDs      []uint64
	TwinIDs          []uint64
	NodeIDs          []uint64
	Types            []string
	States           []string
	Names            []string
	DeploymentDatas  []string
	DeploymentHashes []string

	maxNumberOfPublicIPs uint64
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

func validateContractBillings(local, remote []proxytypes.ContractBilling) error {
	localCp := make([]proxytypes.ContractBilling, len(local))
	remoteCp := make([]proxytypes.ContractBilling, len(remote))
	copy(localCp, local)
	copy(remoteCp, remote)
	sort.Slice(localCp, func(i, j int) bool {
		return localCp[i].Timestamp < localCp[j].Timestamp
	})
	sort.Slice(remoteCp, func(i, j int) bool {
		return remoteCp[i].Timestamp < remoteCp[j].Timestamp
	})
	iter := localCp
	if len(remote) < len(local) {
		iter = remoteCp
	}

	for i := range iter {
		if !reflect.DeepEqual(local[i], remote[i]) {
			return fmt.Errorf("billing %d mismatch: local: %+v, remote: %+v", i, local[i], remote[i])
		}
	}

	if len(local) < len(remote) {
		if len(local) < len(remote) {
			return fmt.Errorf("first in remote after local: %+v", remote[len(local)])
		} else {
			return fmt.Errorf("first in local after remote: %+v", local[len(remote)])
		}
	}

	return nil
}
func validateContractsResults(local, remote []proxytypes.Contract) error {
	iter := local
	if len(remote) < len(local) {
		iter = remote
	}
	for i := range iter {
		localBillings := local[i].Billing
		remoteBillings := remote[i].Billing
		local[i].Billing = nil
		remote[i].Billing = nil
		if !reflect.DeepEqual(local[i], remote[i]) {
			local[i].Billing = localBillings
			remote[i].Billing = remoteBillings
			return fmt.Errorf("contract %d mismatch: local: %+v, remote: %+v", i, local[i], remote[i])
		}
		if err := validateContractBillings(localBillings, remoteBillings); err != nil {
			panic(err)
		}
		local[i].Billing = localBillings
		remote[i].Billing = remoteBillings
	}

	if len(local) < len(remote) {
		if len(local) < len(remote) {
			return fmt.Errorf("first in remote after local: %+v", remote[len(local)])
		} else {
			return fmt.Errorf("first in local after remote: %+v", local[len(remote)])
		}
	}
	return nil
}

func calcContractsAggregates(data *DBData) (res ContractsAggregate) {
	types := make(map[string]struct{})
	for _, contract := range data.nodeContracts {
		res.contractIDs = append(res.contractIDs, contract.contract_id)
		res.maxNumberOfPublicIPs = max(res.maxNumberOfPublicIPs, contract.number_of_public_i_ps)
		res.DeploymentDatas = append(res.DeploymentDatas, contract.deployment_data)
		res.DeploymentHashes = append(res.DeploymentHashes, contract.deployment_hash)
		res.NodeIDs = append(res.NodeIDs, contract.node_id)
		res.States = append(res.States, contract.state)
		res.TwinIDs = append(res.TwinIDs, contract.twin_id)
		types["node"] = struct{}{}
	}

	for typ := range types {
		res.Types = append(res.Types, typ)
	}
	sort.Slice(res.contractIDs, func(i, j int) bool {
		return res.contractIDs[i] < res.contractIDs[j]
	})
	sort.Slice(res.TwinIDs, func(i, j int) bool {
		return res.TwinIDs[i] < res.TwinIDs[j]
	})
	sort.Slice(res.NodeIDs, func(i, j int) bool {
		return res.NodeIDs[i] < res.NodeIDs[j]
	})
	sort.Slice(res.Types, func(i, j int) bool {
		return res.Types[i] < res.Types[j]
	})
	sort.Slice(res.States, func(i, j int) bool {
		return res.States[i] < res.States[j]
	})
	sort.Slice(res.Names, func(i, j int) bool {
		return res.Names[i] < res.Names[j]
	})
	sort.Slice(res.DeploymentDatas, func(i, j int) bool {
		return res.DeploymentDatas[i] < res.DeploymentDatas[j]
	})
	sort.Slice(res.DeploymentHashes, func(i, j int) bool {
		return res.DeploymentHashes[i] < res.DeploymentHashes[j]
	})
	return
}

func randomContractsFilter(agg *ContractsAggregate) proxytypes.ContractFilter {
	var f proxytypes.ContractFilter
	if flip(.05) {
		c := agg.contractIDs[rand.Intn(len(agg.contractIDs))]
		f.ContractID = &c
	}
	if flip(.25) {
		c := agg.TwinIDs[rand.Intn(len(agg.TwinIDs))]
		f.TwinID = &c
	}
	if flip(.25) {
		c := agg.NodeIDs[rand.Intn(len(agg.NodeIDs))]
		f.NodeID = &c
	}
	if flip(.5) {
		c := agg.Types[rand.Intn(len(agg.Types))]
		f.Type = &c
	}
	if flip(.5) {
		c := agg.States[rand.Intn(len(agg.States))]
		f.State = &c
	}
	if flip(.25) && len(agg.Names) != 0 {
		c := agg.Names[rand.Intn(len(agg.Names))]
		f.Name = &c
	}
	if flip(.25) {
		f.NumberOfPublicIps = rndref(0, agg.maxNumberOfPublicIPs)
	}
	if flip(.25) && len(agg.DeploymentDatas) != 0 {
		c := agg.DeploymentDatas[rand.Intn(len(agg.DeploymentDatas))]
		f.DeploymentData = &c
	}
	if flip(.25) && len(agg.DeploymentHashes) != 0 {
		c := agg.DeploymentHashes[rand.Intn(len(agg.DeploymentHashes))]
		f.DeploymentHash = &c
	}
	return f
}

func serializeContractsFilter(f proxytypes.ContractFilter) string {
	res := ""
	if f.ContractID != nil {
		res = fmt.Sprintf("%sContractID: %d\n", res, *f.ContractID)
	}
	if f.TwinID != nil {
		res = fmt.Sprintf("%sTwinID: %d\n", res, *f.TwinID)
	}
	if f.NodeID != nil {
		res = fmt.Sprintf("%sNodeID: %d\n", res, *f.NodeID)
	}
	if f.Type != nil {
		res = fmt.Sprintf("%sType: %s\n", res, *f.Type)
	}
	if f.State != nil {
		res = fmt.Sprintf("%sState: %s\n", res, *f.State)
	}
	if f.Name != nil {
		res = fmt.Sprintf("%sName: %s\n", res, *f.Name)
	}
	if f.NumberOfPublicIps != nil {
		res = fmt.Sprintf("%sNumberOfPublicIps: %d\n", res, *f.NumberOfPublicIps)
	}
	if f.DeploymentData != nil {
		res = fmt.Sprintf("%sDeploymentData: %s\n", res, *f.DeploymentData)
	}
	if f.DeploymentHash != nil {
		res = fmt.Sprintf("%sDeploymentHash: %s\n", res, *f.DeploymentHash)
	}
	return res
}

func contractsPaginationTest(proxyClient, localClient proxyclient.Client) error {
	node := "node"
	f := proxytypes.ContractFilter{
		Type: &node,
	}
	l := proxytypes.Limit{
		Size:     5,
		Page:     1,
		RetCount: true,
	}
	for {
		localContracts, localCount, err := localClient.Contracts(f, l)
		if err != nil {
			return err
		}
		remoteContracts, remoteCount, err := proxyClient.Contracts(f, l)
		if err != nil {
			return err
		}
		if localCount != remoteCount {
			return fmt.Errorf("contracts: node pagination: local count: %d, remote count: %d", localCount, remoteCount)
		}
		if localCount < len(localContracts) {
			return fmt.Errorf("contracts: count in the header %d is less returned length", localCount)
		}
		if remoteCount < len(remoteContracts) {
			return fmt.Errorf("contracts: count in the header %d is less returned length", remoteCount)
		}
		if localCount == 0 {
			fmt.Println("trivial contract pagination test")
		}
		if err := validateContractsResults(localContracts, remoteContracts); err != nil {
			return err
		}
		if l.Page*l.Size >= uint64(localCount) {
			break
		}
		l.Page++
	}
	return nil
}

func ContractsStressTest(data *DBData, proxyClient, localClient proxyclient.Client) error {
	agg := calcContractsAggregates(data)
	for i := 0; i < ContractsTests; i++ {
		l := proxytypes.Limit{
			Size:     999999999999,
			Page:     1,
			RetCount: false,
		}
		f := randomContractsFilter(&agg)

		localContracts, _, err := localClient.Contracts(f, l)
		if err != nil {
			panic(err)
		}
		remoteContracts, _, err := proxyClient.Contracts(f, l)
		if err != nil {
			panic(err)
		}
		contractsReturned[len(remoteContracts)] += 1
		if err := validateContractsResults(localContracts, remoteContracts); err != nil {
			return errors.Wrapf(err, "filter: %s", serializeContractsFilter(f))
		}

	}
	return nil
}

func contractsTest(data *DBData, proxyClient, localClient proxyclient.Client) error {
	if err := contractsPaginationTest(proxyClient, localClient); err != nil {
		panic(err)
	}
	if err := ContractsStressTest(data, proxyClient, localClient); err != nil {
		panic(err)
	}
	keys := make([]int, 0)
	for k, v := range contractsReturned {
		if v != 0 {
			keys = append(keys, k)
		}
	}
	sort.Ints(keys)
	for _, k := range keys {
		fmt.Printf("(%d, %d)", k, contractsReturned[k])
	}
	fmt.Println()
	return nil
}
