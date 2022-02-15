package explorer

import (
	"context"
	"math"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/grid_proxy_server/internal/explorer/db"
	"github.com/threefoldtech/zos/client"
	"github.com/threefoldtech/zos/pkg/rmb"
)

const (
	nodeFetchingPeriod     = 1 * time.Minute
	nodePageSize           = 20
	SSDOverProvisionFactor = 2
)

// NodeRequest for a worker to fetch node information
type NodeRequest struct {
	NodeID int
	TwinID int
}

// NodeManager manages the fetchers
type NodeManager struct {
	db      db.Database
	rmb     rmb.Client
	workers int
	ch      chan NodeRequest
}

// NewNodeManager creates a new manager with number of fetching workers
func NewNodeManager(db db.Database, rmb rmb.Client, workers int) NodeManager {
	return NodeManager{db, rmb, workers, make(chan NodeRequest)}
}

// Run starts the workers and manages fetching requests
func (n *NodeManager) Run(ctx context.Context) {
	n.initWorkers(ctx)
	n.fetchNodes(ctx)
	tc := time.NewTicker(nodeFetchingPeriod)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tc.C:
			n.fetchNodes(ctx)
		}
	}
}

func (n *NodeManager) fetchNodes(ctx context.Context) {
	nodeCursor := db.NewNodeCursor(n.db, nodePageSize)
	for {
		nodes, err := nodeCursor.Next()
		if err != nil {
			log.Error().Err(err).Msg("couldn't get nodes from sqlite db")
			return
		}
		if len(nodes) == 0 {
			return
		}
		for _, node := range nodes {
			if shouldFetch(&node) {
				n.ch <- NodeRequest{
					node.NodeID,
					node.NodeData.TwinID,
				}
			}
		}
	}
}

func shouldFetch(node *db.AllNodeData) bool {
	if node.ConnectionInfo.LastFetchAttempt == 0 {
		// first time
		return true
	}
	power := node.ConnectionInfo.Retries - 3
	if node.ConnectionInfo.Retries <= 3 {
		power = 0
	} else if node.ConnectionInfo.Retries >= 8 {
		power = 5
	}
	waitPeriod := int(math.Pow(2, float64(power)))
	lastFetch := time.Unix(int64(node.ConnectionInfo.LastFetchAttempt), 0)
	nextFetch := lastFetch.Add(time.Duration(time.Duration(waitPeriod) * time.Minute))
	return time.Now().After(nextFetch)
}

func (n *NodeManager) initWorkers(ctx context.Context) {
	i := 0
	for i < n.workers {
		fetcher := NewNodeFetcher(n.db, n.rmb, n.ch)
		go fetcher.Run(ctx)
		i++
	}
}

// NodeFetcher does the actual node data fetching
type NodeFetcher struct {
	db  db.Database
	rmb rmb.Client
	ch  chan NodeRequest
}

// NewNodeFetcher constructs a new node fetcher with a channel receiving its requests on
func NewNodeFetcher(db db.Database, rmb rmb.Client, ch chan NodeRequest) NodeFetcher {
	return NodeFetcher{db, rmb, ch}
}

// Run starts listening on nodes fetching requests
func (n *NodeFetcher) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-n.ch:
			if err := n.fetchNodeData(ctx, &req); err != nil {
				log.Info().Err(err).Int("twin_id", req.TwinID).Int("node_ud", req.NodeID).Msg("couldn't fetch node info")
				if err := n.db.UpdateNodeError(uint32(req.NodeID), err); err != nil {
					log.Error().Err(err).Int("twin_id", req.TwinID).Int("node_ud", req.NodeID).Msg("couldn't update node error")
				}
				continue
			} else {
				log.Debug().Int("twin_id", req.TwinID).Int("node_id", req.NodeID).Msg("node data fetched successfully")
			}
		}
	}
}

func (n *NodeFetcher) fetchNodeData(ctx context.Context, req *NodeRequest) error {
	sub, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cl := client.NewNodeClient(uint32(req.TwinID), n.rmb)
	total, used, err := cl.Counters(sub)
	if err != nil {
		return errors.Wrap(err, "couldn't get node statistics")
	}
	hypervisor, err := cl.SystemHypervisor(ctx)
	if err != nil {
		return errors.Wrap(err, "couldn't get node hyperisor")
	}
	version, err := cl.SystemVersion(ctx)
	if err != nil {
		return errors.Wrap(err, "couldn't get node version")
	}
	nodeInfo := db.PulledNodeData{
		Resources: db.CapacityInfo{
			UsedCRU:   used.CRU,
			FreeSRU:   total.SRU*SSDOverProvisionFactor - used.SRU,
			FreeHRU:   total.HRU - 2,
			FreeMRU:   total.MRU - used.MRU,
			UsedIPV4U: used.IPV4U,
		},
		Status:     "up",
		Hypervisor: hypervisor,
		ZosVersion: version.ZOS,
	}
	if err := n.db.UpdateNodeData(uint32(req.NodeID), nodeInfo); err != nil {
		return err
	}
	return nil
}
