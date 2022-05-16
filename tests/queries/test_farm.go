package main

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/threefoldtech/grid_proxy_server/pkg/gridproxy"
)

var (
	farmsReturned = make(map[int]uint64)
)

const (
	FarmTests = 2000
)

type FarmsAggregate struct {
	stellarAddresses   []string
	pricingPolicyIDs   []uint64
	farmNames          []string
	farmIDs            []uint64
	twinIDs            []uint64
	certificationTypes []string

	maxFreeIPs  uint64
	maxTotalIPs uint64
}

func farmSatisfies(data *DBData, farm farm, f gridproxy.FarmFilter) bool {
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
	if f.Name != nil && *f.Name != farm.name {
		return false
	}
	if f.NameContains != nil && !strings.Contains(farm.name, *f.NameContains) {
		return false
	}
	if f.CertificationType != nil && *f.CertificationType != farm.certification_type {
		return false
	}
	// FIXME: version was removed from the farms
	/*
		TODO:
			Dedicated         *bool
	*/
	return true
}

func validatePublicIPs(local, remote []gridproxy.PublicIP) error {
	localIPs := make(map[string]gridproxy.PublicIP)
	remoteIPs := make(map[string]gridproxy.PublicIP)
	for _, ip := range local {
		localIPs[ip.ID] = ip
	}
	for _, ip := range remote {
		remoteIPs[ip.ID] = ip
	}
	for _, ip := range remote {
		if _, ok := localIPs[ip.ID]; !ok {
			return fmt.Errorf("ip %s exists in remote but not in local", ip.ID)
		}
		if !reflect.DeepEqual(localIPs[ip.ID], remoteIPs[ip.ID]) {
			return fmt.Errorf("ip %s mismatch: local: %+v, remote: %+v", ip.ID, localIPs[ip.ID], remoteIPs[ip.ID])
		}
	}
	for _, ip := range local {
		if _, ok := localIPs[ip.ID]; !ok {
			return fmt.Errorf("ip %s exists in local but not in remote", ip.ID)
		}
		if !reflect.DeepEqual(localIPs[ip.ID], remoteIPs[ip.ID]) {
			return fmt.Errorf("ip %s mismatch: local: %+v, remote: %+v", ip.ID, localIPs[ip.ID], remoteIPs[ip.ID])
		}
	}
	return nil
}

func validateFarmsResults(local, remote []gridproxy.Farm) error {
	iter := local
	if len(remote) < len(local) {
		iter = remote
	}
	for i := range iter {
		localIPs := local[i].PublicIps
		remoteIPs := remote[i].PublicIps
		if err := validatePublicIPs(localIPs, remoteIPs); err != nil {
			return err
		}
		local[i].PublicIps = nil
		remote[i].PublicIps = nil
		if !reflect.DeepEqual(local[i], remote[i]) {
			local[i].PublicIps = localIPs
			remote[i].PublicIps = remoteIPs
			return fmt.Errorf("farm %d mismatch: local: %+v, remote: %+v", i, local[i], remote[i])
		}
		local[i].PublicIps = localIPs
		remote[i].PublicIps = remoteIPs
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

func calcFarmsAggregates(data *DBData) (res FarmsAggregate) {
	for _, farm := range data.farms {
		res.farmNames = append(res.farmNames, farm.name)
		res.stellarAddresses = append(res.farmNames, farm.stellar_address)
		res.pricingPolicyIDs = append(res.pricingPolicyIDs, farm.pricing_policy_id)
		res.certificationTypes = append(res.certificationTypes, farm.certification_type)
		res.farmIDs = append(res.farmIDs, farm.farm_id)
		res.twinIDs = append(res.twinIDs, farm.twin_id)
	}

	farmIPs := make(map[uint64]uint64)
	farmTotalIPs := make(map[uint64]uint64)
	for _, publicIP := range data.publicIPs {
		if publicIP.contract_id == 0 {
			farmIPs[data.farmIDMap[publicIP.farm_id]] += 1
		}
		farmTotalIPs[data.farmIDMap[publicIP.farm_id]] += 1
	}
	for _, cnt := range farmIPs {
		res.maxFreeIPs = max(res.maxFreeIPs, cnt)
	}
	for _, cnt := range farmTotalIPs {
		res.maxTotalIPs = max(res.maxTotalIPs, cnt)
	}
	return
}

func randomFarmsFilter(agg *FarmsAggregate) gridproxy.FarmFilter {
	var f gridproxy.FarmFilter
	if flip(.5) {
		f.FreeIPs = rndref(0, agg.maxFreeIPs)
	}
	if flip(.5) {
		f.TotalIPs = rndref(0, agg.maxTotalIPs)
	}
	if flip(.05) {
		c := agg.stellarAddresses[rand.Intn(len(agg.stellarAddresses))]
		f.StellarAddress = &c
	}
	if flip(.5) {
		c := agg.pricingPolicyIDs[rand.Intn(len(agg.pricingPolicyIDs))]
		f.PricingPolicyID = &c
	}
	if flip(.05) {
		c := agg.farmIDs[rand.Intn(len(agg.farmIDs))]
		f.FarmID = &c
	}
	if flip(.05) {
		c := agg.twinIDs[rand.Intn(len(agg.twinIDs))]
		f.TwinID = &c
	}
	if flip(.05) {
		c := agg.farmNames[rand.Intn(len(agg.farmNames))]
		f.Name = &c
	}
	if flip(.05) {
		c := agg.farmNames[rand.Intn(len(agg.farmNames))]
		a, b := rand.Intn(len(c)), rand.Intn(len(c))
		if a > b {
			a, b = b, a
		}
		c = c[a : b+1]
		f.NameContains = &c
	}
	if flip(.5) {
		c := agg.certificationTypes[rand.Intn(len(agg.certificationTypes))]
		f.CertificationType = &c
	}
	// Dedicated         *bool

	return f
}

func serializeFarmsFilter(f gridproxy.FarmFilter) string {
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
		res = fmt.Sprintf("%sCertificationType: %s\n", res, *f.CertificationType)
	}
	/*
		TODO:
			Dedicated         *bool
	*/
	return res
}

func FarmsStressTest(data *DBData, proxyClient, localClient gridproxy.GridProxyClient) error {
	agg := calcFarmsAggregates(data)
	for i := 0; i < Tests; i++ {
		l := gridproxy.Limit{
			Size:     999999999999,
			Page:     1,
			RetCount: false,
		}
		f := randomFarmsFilter(&agg)

		localFarms, err := localClient.Farms(f, l)
		if err != nil {
			return err
		}
		remoteFarms, err := proxyClient.Farms(f, l)
		if err != nil {
			return err
		}
		farmsReturned[len(remoteFarms)] += 1
		if err := validateFarmsResults(localFarms, remoteFarms); err != nil {
			return errors.Wrapf(err, "filter: %s", serializeFarmsFilter(f))
		}

	}
	return nil
}

func FarmsTest(data *DBData, proxyClient, localClient gridproxy.GridProxyClient) error {
	if err := FarmsStressTest(data, proxyClient, localClient); err != nil {
		return err
	}
	keys := make([]int, 0)
	for k, v := range farmsReturned {
		if v != 0 {
			keys = append(keys, k)
		}
	}
	sort.Ints(keys)
	for _, k := range keys {
		fmt.Printf("(%d, %d)", k, farmsReturned[k])
	}
	fmt.Println()
	return nil
}
