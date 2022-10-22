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

type NodesAggregate struct {
	countries []string
	cities    []string
	farmNames []string
	farmIDs   []uint64
	freeMRUs  []uint64
	freeSRUs  []uint64
	freeHRUs  []uint64

	maxFreeMRU  uint64
	maxFreeSRU  uint64
	maxFreeHRU  uint64
	maxFreeIPs  uint64
	nodeRenters []uint64
	twins       []uint64
}

var (
	NODE_COUNT      = 1000
	NODE_TESTS      = 2000
	ErrNodeNotFound = errors.New("node not found")
)

func TestNode(t *testing.T) {
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

	t.Run("node pagination test", func(t *testing.T) {
		nodePaginationCheck(t, localClient, proxyClient)
	})

	t.Run("single node test", func(t *testing.T) {
		singleNodeCheck(t, localClient, proxyClient)
	})

	t.Run("node up test", func(t *testing.T) {
		f := proxytypes.NodeFilter{
			Status: &STATUS_UP,
		}
		l := proxytypes.Limit{
			Size:     999999999,
			Page:     1,
			RetCount: true,
		}
		localNodes, _, err := localClient.Nodes(f, l)
		assert.NoError(t, err)
		remoteNodes, _, err := proxyClient.Nodes(f, l)
		assert.NoError(t, err)
		assert.True(t, reflect.DeepEqual(localNodes, remoteNodes))
	})

	t.Run("node status test", func(t *testing.T) {
		for i := 1; i <= NODE_COUNT; i++ {
			if flip(.3) {
				localNodeStatus, err := localClient.NodeStatus(uint32(i))
				assert.NoError(t, err)
				remoteNodeStatus, err := proxyClient.NodeStatus(uint32(i))
				assert.NoError(t, err)
				assert.True(t, reflect.DeepEqual(localNodeStatus, remoteNodeStatus))
			}
		}
	})

	t.Run("node stress test", func(t *testing.T) {
		agg := calcNodesAggregates(&data)
		for i := 0; i < NODE_TESTS; i++ {
			l := proxytypes.Limit{
				Size:     999999999999,
				Page:     1,
				RetCount: false,
			}
			f := randomNodeFilter(&agg)
			localNodes, _, err := localClient.Nodes(f, l)
			assert.NoError(t, err)
			remoteNodes, _, err := proxyClient.Nodes(f, l)
			assert.NoError(t, err)
			assert.Equal(t, len(localNodes), len(remoteNodes))
			if len(localNodes) != len(remoteNodes) {
				continue
			}
			for id := 0; id < len(localNodes); id++ {
				assert.True(t, reflect.DeepEqual(localNodes[id], remoteNodes[id]))
			}
		}
	})

	t.Run("node not found test", func(t *testing.T) {
		nodeID := 1000000000
		_, err := proxyClient.Node(uint32(nodeID))
		assert.Equal(t, err.Error(), ErrNodeNotFound.Error())
	})

	t.Run("nodes test without resources view", func(t *testing.T) {
		db := data.db
		_, err := db.Exec("drop view nodes_resources_view ;")
		assert.NoError(t, err)
		singleNodeCheck(t, localClient, proxyClient)
		assert.NoError(t, err)
		_, err = db.Exec("drop view nodes_resources_view ;")
		assert.NoError(t, err)
		nodePaginationCheck(t, localClient, proxyClient)
	})
}

func singleNodeCheck(t *testing.T, localClient proxyclient.Client, proxyClient proxyclient.Client) {
	nodeID := rand.Intn(NODE_COUNT)
	localNode, err := localClient.Node(uint32(nodeID))
	assert.NoError(t, err)
	remoteNode, err := proxyClient.Node(uint32(nodeID))
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(localNode, remoteNode))
}

func nodePaginationCheck(t *testing.T, localClient proxyclient.Client, proxyClient proxyclient.Client) {
	f := proxytypes.NodeFilter{
		Status: &STATUS_DOWN,
	}
	l := proxytypes.Limit{
		Size:     5,
		Page:     1,
		RetCount: true,
	}
	for ; ; l.Page++ {
		localNodes, localCount, err := localClient.Nodes(f, l)
		assert.NoError(t, err)
		remoteNodes, remoteCount, err := proxyClient.Nodes(f, l)
		assert.NoError(t, err)
		assert.Equal(t, remoteCount, localCount, "local and remote counts are not equal")
		assert.True(t, reflect.DeepEqual(localNodes, remoteNodes))
		if l.Page*l.Size >= uint64(localCount) {
			break
		}
	}
}

func randomNodeFilter(agg *NodesAggregate) proxytypes.NodeFilter {
	var f proxytypes.NodeFilter
	if flip(.5) { // status
		status := "down"
		if flip(.5) {
			status = "up"
		}
		f.Status = &status
	}
	if flip(.5) {
		if flip(.1) {
			c := agg.freeMRUs[rand.Intn(len(agg.freeMRUs))]
			f.FreeMRU = &c
		} else {
			f.FreeMRU = rndref(0, agg.maxFreeMRU)
		}
	}
	if flip(.5) {
		if flip(.1) {
			c := agg.freeHRUs[rand.Intn(len(agg.freeHRUs))]
			f.FreeHRU = &c
		} else {
			f.FreeHRU = rndref(0, agg.maxFreeHRU)
		}
	}
	if flip(.5) {
		if flip(.1) {
			c := agg.freeSRUs[rand.Intn(len(agg.freeSRUs))]
			f.FreeSRU = &c
		} else {
			f.FreeSRU = rndref(0, agg.maxFreeSRU)
		}
	}
	if flip(.05) {
		c := agg.countries[rand.Intn(len(agg.countries))]
		f.Country = &c
	}
	if flip(.05) {
		c := agg.farmNames[rand.Intn(len(agg.farmNames))]
		f.FarmName = &c
	}
	if flip(.05) {
		for _, id := range agg.farmIDs {
			if flip(float32(min(3, uint64(len(agg.farmIDs)))) / float32(len(agg.farmIDs))) {
				f.FarmIDs = append(f.FarmIDs, id)
			}
		}
	}
	if flip(.05) {
		f.FreeIPs = rndref(0, agg.maxFreeIPs)
	}
	if flip(.05) {
		v := true
		// if flip(.5) {
		// 	v = false
		// }
		f.IPv4 = &v
	}
	if flip(.05) {
		v := true
		// if flip(.5) {
		// 	v = false
		// }
		f.IPv6 = &v
	}
	if flip(.05) {
		v := true
		// currently, it's not checked against
		// if flip(.5) {
		// 	v = false
		// }
		f.Domain = &v
	}
	if flip(.5) {
		v := uint64(rand.Intn(1100)) // 1000 is the total nodes + 100 for non-existed cases
		f.NodeID = &v
	}
	if flip(.5) {
		v := uint64(rand.Intn(3500))
		f.TwinID = &v
	}
	if flip(.05) {
		v := true
		if flip(.5) {
			v = false
		}
		f.Rentable = &v
	}
	if flip(.3) {
		c := agg.twins[rand.Intn(len(agg.twins))]
		if flip(.9) && len(agg.nodeRenters) != 0 {
			c = agg.nodeRenters[rand.Intn(len(agg.nodeRenters))]
		}
		f.RentedBy = &c
	}
	if flip(.3) {
		c := agg.twins[rand.Intn(len(agg.twins))]
		if flip(.1) && len(agg.nodeRenters) != 0 {
			c = agg.nodeRenters[rand.Intn(len(agg.nodeRenters))]
		}
		f.AvailableFor = &c
	}
	if flip(.1) {
		v := true
		if flip(.5) {
			v = false
		}
		f.Rented = &v
	}
	return f
}

func calcNodesAggregates(data *DBData) (res NodesAggregate) {
	cities := make(map[string]struct{})
	countries := make(map[string]struct{})
	for _, node := range data.nodes {
		cities[node.city] = struct{}{}
		countries[node.country] = struct{}{}
		free := calcFreeResources(data.nodeTotalResources[node.node_id], data.nodeUsedResources[node.node_id])
		res.maxFreeHRU = max(res.maxFreeHRU, free.hru)
		res.maxFreeSRU = max(res.maxFreeSRU, free.sru)
		res.maxFreeMRU = max(res.maxFreeMRU, free.mru)
		res.freeMRUs = append(res.freeMRUs, free.mru)
		res.freeSRUs = append(res.freeSRUs, free.sru)
		res.freeHRUs = append(res.freeHRUs, free.hru)
	}
	for _, contract := range data.rentContracts {
		if contract.state == "Deleted" {
			continue
		}
		res.nodeRenters = append(res.nodeRenters, contract.twin_id)
	}
	for _, twin := range data.twins {
		res.twins = append(res.twins, twin.twin_id)
	}
	for city := range cities {
		res.cities = append(res.cities, city)
	}
	for country := range countries {
		res.countries = append(res.cities, country)
	}
	for _, farm := range data.farms {
		res.farmNames = append(res.farmNames, farm.name)
		res.farmIDs = append(res.farmIDs, farm.farm_id)
	}

	farmIPs := make(map[uint64]uint64)
	for _, publicIP := range data.publicIPs {
		if publicIP.contract_id == 0 {
			farmIPs[data.farmIDMap[publicIP.farm_id]] += 1
		}
	}
	for _, cnt := range farmIPs {
		res.maxFreeIPs = max(res.maxFreeIPs, cnt)
	}
	sort.Slice(res.countries, func(i, j int) bool {
		return res.countries[i] < res.countries[j]
	})
	sort.Slice(res.cities, func(i, j int) bool {
		return res.cities[i] < res.cities[j]
	})
	sort.Slice(res.farmNames, func(i, j int) bool {
		return res.farmNames[i] < res.farmNames[j]
	})
	sort.Slice(res.farmIDs, func(i, j int) bool {
		return res.farmIDs[i] < res.farmIDs[j]
	})
	sort.Slice(res.freeMRUs, func(i, j int) bool {
		return res.freeMRUs[i] < res.freeMRUs[j]
	})
	sort.Slice(res.freeSRUs, func(i, j int) bool {
		return res.freeSRUs[i] < res.freeSRUs[j]
	})
	sort.Slice(res.freeHRUs, func(i, j int) bool {
		return res.freeHRUs[i] < res.freeHRUs[j]
	})
	sort.Slice(res.nodeRenters, func(i, j int) bool {
		return res.nodeRenters[i] < res.nodeRenters[j]
	})
	sort.Slice(res.twins, func(i, j int) bool {
		return res.twins[i] < res.twins[j]
	})
	return
}
