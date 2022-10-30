package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	proxyclient "github.com/threefoldtech/grid_proxy_server/pkg/client"
	proxytypes "github.com/threefoldtech/grid_proxy_server/pkg/types"
)

const (
	FARM_TESTS = 2000
)

type FarmsAggregate struct {
	stellarAddresses []string
	pricingPolicyIDs []uint64
	farmNames        []string
	farmIDs          []uint64
	twinIDs          []uint64
	certifications   []string

	maxFreeIPs  uint64
	maxTotalIPs uint64
}

func TestFarm(t *testing.T) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, POSTGRES_PASSSWORD, POSTGRES_DB)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(errors.Wrap(err, "failed to open db"))
	}
	defer db.Close()

	data, err := load(db)
	if err != nil {
		panic(err)
	}
	proxyClient := proxyclient.NewClient(ENDPOINT)
	localClient := NewGridProxyClient(data)

	t.Run("farms pagination test", func(t *testing.T) {
		one := uint64(1)
		f := proxytypes.FarmFilter{
			TotalIPs: &one,
		}
		l := proxytypes.Limit{
			Size:     5,
			Page:     1,
			RetCount: true,
		}
		for ; ; l.Page++ {
			localFarms, localCount, err := localClient.Farms(f, l)
			assert.NoError(t, err)
			remoteFarms, remoteCount, err := proxyClient.Farms(f, l)
			assert.NoError(t, err)
			assert.Equal(t, localCount, remoteCount)
			err = validateFarmsResults(localFarms, remoteFarms)
			assert.NoError(t, err, serializeFarmsFilter(f))
			if l.Page*l.Size >= uint64(localCount) {
				break
			}
		}
	})

	t.Run("farms stress test", func(t *testing.T) {
		agg := calcFarmsAggregates(&data)
		for i := 0; i < FARM_TESTS; i++ {
			l := proxytypes.Limit{
				Size:     999999999999,
				Page:     1,
				RetCount: false,
			}
			f := randomFarmsFilter(&agg)
			localFarms, _, err := localClient.Farms(f, l)
			assert.NoError(t, err)
			remoteFarms, _, err := proxyClient.Farms(f, l)
			assert.NoError(t, err)

			err = validateFarmsResults(localFarms, remoteFarms)
			assert.NoError(t, err, serializeFarmsFilter(f))
		}
	})
}

func calcFarmsAggregates(data *DBData) (res FarmsAggregate) {
	for _, farm := range data.farms {
		res.farmNames = append(res.farmNames, farm.name)
		res.stellarAddresses = append(res.stellarAddresses, farm.stellar_address)
		res.pricingPolicyIDs = append(res.pricingPolicyIDs, farm.pricing_policy_id)
		res.certifications = append(res.certifications, farm.certification)
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

	sort.Slice(res.stellarAddresses, func(i, j int) bool {
		return res.stellarAddresses[i] < res.stellarAddresses[j]
	})
	sort.Slice(res.pricingPolicyIDs, func(i, j int) bool {
		return res.pricingPolicyIDs[i] < res.pricingPolicyIDs[j]
	})
	sort.Slice(res.farmNames, func(i, j int) bool {
		return res.farmNames[i] < res.farmNames[j]
	})
	sort.Slice(res.farmIDs, func(i, j int) bool {
		return res.farmIDs[i] < res.farmIDs[j]
	})
	sort.Slice(res.twinIDs, func(i, j int) bool {
		return res.twinIDs[i] < res.twinIDs[j]
	})
	sort.Slice(res.certifications, func(i, j int) bool {
		return res.certifications[i] < res.certifications[j]
	})

	return
}

func randomFarmsFilter(agg *FarmsAggregate) proxytypes.FarmFilter {
	var f proxytypes.FarmFilter
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
		v := changeCase(c)
		f.Name = &v
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
		c := agg.certifications[rand.Intn(len(agg.certifications))]
		f.CertificationType = &c
	}
	if flip(.5) {
		v := true
		if flip(.5) {
			v = false
		}
		f.Dedicated = &v
	}

	return f
}

func validateFarmsResults(local, remote []proxytypes.Farm) error {
	iter := local
	if len(remote) < len(local) {
		iter = remote
	}
	for i := range iter {
		localIPs := local[i].PublicIps
		remoteIPs := remote[i].PublicIps
		local[i].PublicIps = nil
		remote[i].PublicIps = nil
		if !reflect.DeepEqual(local[i], remote[i]) {
			local[i].PublicIps = localIPs
			remote[i].PublicIps = remoteIPs
			return fmt.Errorf("farm %d mismatch: local: %+v, remote: %+v", i, local[i], remote[i])
		}
		if err := validatePublicIPs(localIPs, remoteIPs); err != nil {
			return err
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

func validatePublicIPs(local, remote []proxytypes.PublicIP) error {
	localIPs := make(map[string]proxytypes.PublicIP)
	remoteIPs := make(map[string]proxytypes.PublicIP)
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
