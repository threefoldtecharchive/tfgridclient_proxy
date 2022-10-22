package main

import (
	"fmt"
	"math/rand"
	"time"

	proxytypes "github.com/threefoldtech/grid_proxy_server/pkg/types"
)

var (
	nodeStateFactor int64 = 3
	reportInterval        = time.Hour
)

func calcFreeResources(total node_resources_total, used node_resources_total) node_resources_total {
	if total.mru < used.mru {
		panic("total mru is less than mru")
	}
	if total.hru < used.hru {
		panic("total hru is less than hru")
	}
	if total.sru < used.sru {
		panic("total sru is less than sru")
	}
	return node_resources_total{
		hru: total.hru - used.hru,
		sru: total.sru - used.sru,
		mru: total.mru - used.mru,
	}
}

func isIn(l []uint64, v uint64) bool {
	for _, i := range l {
		if i == v {
			return true
		}
	}
	return false
}

func isUp(nodeID uint64, cache map[uint64]node_status_cache, timestamp uint64) bool {
	status := cache[nodeID].status
	if status == "up" || status == "down" {
		return status == "up"
	}
	// log.Printf("nodeid: %d has no status cache", nodeID)
	return int64(timestamp) > time.Now().Unix()*1000-nodeStateFactor*int64(reportInterval/time.Millisecond)
}

func flip(success float32) bool {
	return rand.Float32() < success
}

func rndref(min, max uint64) *uint64 {
	v := rand.Uint64()%(max-min+1) + min
	return &v
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}
func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func serializeTwinsFilter(f proxytypes.TwinFilter) string {
	res := ""
	if f.TwinID != nil {
		res = fmt.Sprintf("%sTwinID: %d\n", res, *f.TwinID)
	}
	if f.AccountID != nil {
		res = fmt.Sprintf("%sAccountID: %s\n", res, *f.AccountID)
	}
	return res
}

func serializeFilter(f proxytypes.NodeFilter) string {
	res := ""
	if f.Status != nil {
		res = fmt.Sprintf("%sstatus: %s\n", res, *f.Status)
	}
	if f.FreeMRU != nil {
		res = fmt.Sprintf("%sFreeMRU: %d\n", res, *f.FreeMRU)
	}
	if f.FreeSRU != nil {
		res = fmt.Sprintf("%sFreeSRU: %d\n", res, *f.FreeSRU)
	}
	if f.FreeHRU != nil {
		res = fmt.Sprintf("%sFreeHRU: %d\n", res, *f.FreeHRU)
	}
	if f.Country != nil {
		res = fmt.Sprintf("%sCountry: %s\n", res, *f.Country)
	}
	if f.City != nil {
		res = fmt.Sprintf("%sCity: %s\n", res, *f.City)
	}
	if f.FarmName != nil {
		res = fmt.Sprintf("%sFarmName: %s\n", res, *f.FarmName)
	}
	if f.FarmIDs != nil {
		res = fmt.Sprintf("%sFarmIDs: %v\n", res, f.FarmIDs)
	}
	if f.FreeIPs != nil {
		res = fmt.Sprintf("%sFreeIPs: %d\n", res, *f.FreeIPs)
	}
	if f.IPv4 != nil {
		res = fmt.Sprintf("%sIPv4: %t\n", res, *f.IPv4)
	}
	if f.IPv6 != nil {
		res = fmt.Sprintf("%sIPv6: %t\n", res, *f.IPv6)
	}
	if f.Domain != nil {
		res = fmt.Sprintf("%sDomain: %t\n", res, *f.Domain)
	}
	if f.Rentable != nil {
		res = fmt.Sprintf("%sRentable: %t\n", res, *f.Rentable)
	}
	if f.Rentable != nil {
		res = fmt.Sprintf("%sRentable: %t\n", res, *f.Rentable)
	}
	if f.AvailableFor != nil {
		res = fmt.Sprintf("%sAvailableFor: %d\n", res, *f.AvailableFor)
	}
	if f.Rented != nil {
		res = fmt.Sprintf("%sRented: %t\n", res, *f.Rented)
	}
	return res
}

func serializeFarmsFilter(f proxytypes.FarmFilter) string {
	res := ""
	if f.FreeIPs != nil {
		res = fmt.Sprintf("%sFreeIPs: %d\n", res, *f.FreeIPs)
	}
	if f.TotalIPs != nil {
		res = fmt.Sprintf("%sTotalIPs: %d\n", res, *f.TotalIPs)
	}
	if f.StellarAddress != nil {
		res = fmt.Sprintf("%sStellarAddress: %s\n", res, *f.StellarAddress)
	}
	if f.PricingPolicyID != nil {
		res = fmt.Sprintf("%sPricingPolicyID: %d\n", res, *f.PricingPolicyID)
	}
	if f.FarmID != nil {
		res = fmt.Sprintf("%sFarmID: %d\n", res, *f.FarmID)
	}
	if f.TwinID != nil {
		res = fmt.Sprintf("%sTwinID: %d\n", res, *f.TwinID)
	}
	if f.Name != nil {
		res = fmt.Sprintf("%sName: %s\n", res, *f.Name)
	}
	if f.NameContains != nil {
		res = fmt.Sprintf("%sNameContains: %s\n", res, *f.NameContains)
	}
	if f.CertificationType != nil {
		res = fmt.Sprintf("%sCertification: %s\n", res, *f.CertificationType)
	}
	if f.Dedicated != nil {
		res = fmt.Sprintf("%sDedicated: %t\n", res, *f.Dedicated)
	}
	return res
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
