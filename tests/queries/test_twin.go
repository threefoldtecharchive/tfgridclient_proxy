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
	twinsReturned = make(map[int]uint64)
)

const (
	TwinsTests = 200
)

type TwinsAggregate struct {
	twinIDs    []uint64
	accountIDs []string
	ip         []string
	twins      map[uint64]twin
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

func validateTwinsResults(local, remote []proxytypes.Twin) error {
	iter := local
	if len(remote) < len(local) {
		iter = remote
	}
	for i := range iter {
		if !reflect.DeepEqual(local[i], remote[i]) {
			return fmt.Errorf("twin %d mismatch: local: %+v, remote: %+v", i, local[i], remote[i])
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

func calcTwinsAggregates(data *DBData) (res TwinsAggregate) {
	for _, twin := range data.twins {
		res.twinIDs = append(res.twinIDs, twin.twin_id)
		res.accountIDs = append(res.accountIDs, twin.account_id)
		res.ip = append(res.ip, twin.ip)
	}
	res.twins = data.twins
	sort.Slice(res.twinIDs, func(i, j int) bool {
		return res.twinIDs[i] < res.twinIDs[j]
	})
	sort.Slice(res.accountIDs, func(i, j int) bool {
		return res.accountIDs[i] < res.accountIDs[j]
	})
	sort.Slice(res.ip, func(i, j int) bool {
		return res.ip[i] < res.ip[j]
	})
	return
}

func randomTwinsFilter(agg *TwinsAggregate) proxytypes.TwinFilter {
	var f proxytypes.TwinFilter
	if flip(.2) {
		c := agg.twinIDs[rand.Intn(len(agg.twinIDs))]
		f.TwinID = &c
	}
	if flip(.2) {
		if f.TwinID != nil && flip(.4) {
			accountID := agg.twins[*f.TwinID].account_id
			f.AccountID = &accountID
		} else {
			c := agg.accountIDs[rand.Intn(len(agg.accountIDs))]
			f.AccountID = &c
		}
	}

	return f
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

func twinsPaginationTest(proxyClient, localClient proxyclient.Client) error {
	f := proxytypes.TwinFilter{}
	l := proxytypes.Limit{
		Size:     5,
		Page:     1,
		RetCount: true,
	}
	for {
		localTwins, localCount, err := localClient.Twins(f, l)
		if err != nil {
			return err
		}
		remoteTwins, remoteCount, err := proxyClient.Twins(f, l)
		if err != nil {
			return err
		}
		if localCount != remoteCount {
			return fmt.Errorf("twins: node pagination: local count: %d, remote count: %d", localCount, remoteCount)
		}
		if localCount < len(localTwins) {
			return fmt.Errorf("twins: count in the header %d is less returned length", localCount)
		}
		if remoteCount < len(remoteTwins) {
			return fmt.Errorf("twins: count in the header %d is less returned length", remoteCount)
		}
		if localCount == 0 {
			fmt.Println("trivial twin pagination test")
		}
		if err := validateTwinsResults(localTwins, remoteTwins); err != nil {
			return err
		}
		if l.Page*l.Size >= uint64(localCount) {
			break
		}
		l.Page++
	}
	return nil
}

func TwinsStressTest(data *DBData, proxyClient, localClient proxyclient.Client) error {
	agg := calcTwinsAggregates(data)
	for i := 0; i < TwinsTests; i++ {
		l := proxytypes.Limit{
			Size:     999999999999,
			Page:     1,
			RetCount: false,
		}
		f := randomTwinsFilter(&agg)

		localTwins, _, err := localClient.Twins(f, l)
		if err != nil {
			panic(err)
		}
		remoteTwins, _, err := proxyClient.Twins(f, l)
		if err != nil {
			panic(err)
		}
		twinsReturned[len(remoteTwins)] += 1
		if err := validateTwinsResults(localTwins, remoteTwins); err != nil {
			return errors.Wrapf(err, "filter: %s", serializeTwinsFilter(f))
		}

	}
	return nil
}

func twinsTest(data *DBData, proxyClient, localClient proxyclient.Client) error {
	if err := twinsPaginationTest(proxyClient, localClient); err != nil {
		panic(err)
	}
	if err := TwinsStressTest(data, proxyClient, localClient); err != nil {
		panic(err)
	}
	keys := make([]int, 0)
	for k, v := range twinsReturned {
		if v != 0 {
			keys = append(keys, k)
		}
	}
	sort.Ints(keys)
	for _, k := range keys {
		fmt.Printf("(%d, %d)", k, twinsReturned[k])
	}
	fmt.Println()
	return nil
}
