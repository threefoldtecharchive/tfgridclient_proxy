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

type TwinsAggregate struct {
	twinIDs    []uint64
	accountIDs []string
	relays     []string
	publicKeys []string
	twins      map[uint64]twin
}

const (
	TWINS_TESTS = 200
)

func TestTwins(t *testing.T) {
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

	t.Run("twins pagination test", func(t *testing.T) {
		f := proxytypes.TwinFilter{}
		l := proxytypes.Limit{
			Size:     5,
			Page:     1,
			RetCount: true,
		}
		for {
			localTwins, localCount, err := localClient.Twins(f, l)
			assert.NoError(t, err)
			remoteTwins, remoteCount, err := proxyClient.Twins(f, l)
			assert.NoError(t, err)
			assert.Equal(t, localCount, remoteCount)
			err = validateTwinsResults(localTwins, remoteTwins)
			assert.NoError(t, err)
			if l.Page*l.Size >= uint64(localCount) {
				break
			}
			l.Page++
		}
	})

	t.Run("twins stress test", func(t *testing.T) {
		agg := calcTwinsAggregates(&data)
		for i := 0; i < TWINS_TESTS; i++ {
			l := proxytypes.Limit{
				Size:     999999999999,
				Page:     1,
				RetCount: false,
			}
			f := randomTwinsFilter(&agg)
			localTwins, _, err := localClient.Twins(f, l)
			assert.NoError(t, err)
			remoteTwins, _, err := proxyClient.Twins(f, l)
			assert.NoError(t, err)
			err = validateTwinsResults(localTwins, remoteTwins)
			assert.NoError(t, err, serializeTwinsFilter(f))
		}
	})
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
	if flip(.2) {
		if f.TwinID != nil && flip(.4) {
			relay := agg.twins[*f.TwinID].account_id
			f.Relay = &relay
		} else {
			c := agg.relays[rand.Intn(len(agg.relays))]
			f.Relay = &c
		}
	}
	if flip(.2) {
		if f.TwinID != nil && flip(.4) {
			publicKey := agg.twins[*f.TwinID].account_id
			f.PublicKey = &publicKey
		} else {
			c := agg.publicKeys[rand.Intn(len(agg.publicKeys))]
			f.PublicKey = &c
		}
	}

	return f
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
		res.relays = append(res.relays, twin.relay)
		res.publicKeys = append(res.publicKeys, twin.public_key)
	}
	res.twins = data.twins
	sort.Slice(res.twinIDs, func(i, j int) bool {
		return res.twinIDs[i] < res.twinIDs[j]
	})
	sort.Slice(res.accountIDs, func(i, j int) bool {
		return res.accountIDs[i] < res.accountIDs[j]
	})
	sort.Slice(res.relays, func(i, j int) bool {
		return res.relays[i] < res.relays[j]
	})
	sort.Slice(res.publicKeys, func(i, j int) bool {
		return res.publicKeys[i] < res.publicKeys[j]
	})
	return
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
