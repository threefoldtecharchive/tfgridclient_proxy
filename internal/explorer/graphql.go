package explorer

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/grid_proxy_server/internal/explorer/db"
)

const (
	QueryLimit     = 20
	FarmSyncPeriod = 5 * time.Minute
	NodeSyncPeriod = 30 * time.Minute
)

type NodeSyncer struct {
	client *GraphqlClient
	db     db.Database
}

type FarmSyncer struct {
	client *GraphqlClient
	db     db.Database
}

func NewFarmSyncer(client *GraphqlClient, db db.Database) FarmSyncer {
	return FarmSyncer{client, db}
}
func NewNodeSyncer(client *GraphqlClient, db db.Database) NodeSyncer {
	return NodeSyncer{client, db}
}

func (fs *NodeSyncer) Run(ctx context.Context) {
	fs.syncNodes(ctx)
	tc := time.NewTicker(NodeSyncPeriod)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tc.C:
			fs.syncNodes(ctx)
		}
	}
}

func (fs *NodeSyncer) syncNodes(ctx context.Context) {
	nodeCursor := NewNodeCursor(fs.client, QueryLimit)
	for {
		nodes, err := nodeCursor.Next()
		if err != nil {
			log.Error().Err(err).Msg("failed to get graphql farms")
			return
		}
		if len(nodes) == 0 {
			return
		}
		for _, node := range nodes {
			if err := fs.db.InsertOrUpdateNodeGraphqlData(uint32(node.NodeID), node); err != nil {
				log.Error().Err(err).Int("node_id", node.NodeID).Msg("failed to update node in db")
			}
		}
	}
}

func (fs *FarmSyncer) Run(ctx context.Context) {
	fs.syncFarms(ctx)
	tc := time.NewTicker(FarmSyncPeriod)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tc.C:
			fs.syncFarms(ctx)
		}
	}
}

func (fs *FarmSyncer) syncFarms(ctx context.Context) {
	farmCursor := NewFarmCursor(fs.client, QueryLimit)
	for {
		farms, err := farmCursor.Next()
		if err != nil {
			log.Error().Err(err).Msg("failed to get graphql farms")
			return
		}
		if len(farms) == 0 {
			return
		}
		for _, farm := range farms {
			if err := fs.db.UpdateFarm(farm); err != nil {
				log.Error().Err(err).Int("farm_id", farm.FarmID).Msg("failed to update farm in db")
			}
		}
	}
}
