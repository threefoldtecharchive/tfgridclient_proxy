package explorer

import (
	"context"

	"github.com/gomodule/redigo/redis"
	"github.com/threefoldtech/zos/pkg/capacity/dmi"
	"github.com/threefoldtech/zos/pkg/gridtypes"
	"github.com/threefoldtech/zos/pkg/rmb"
)

// URL is the default explorer graphql url
const URL string = "https://explorer.devnet.grid.tf/graphql/"

// App is the main app objects
type App struct {
	debug    bool
	explorer string
	redis    *redis.Pool
	ctx      context.Context
	rmb      rmb.Client
}

// OffsetKey is the type holds the request context
type OffsetKey struct{}

// SpecificFarmKey is the type holds the request context
type SpecificFarmKey struct{}

// MaxResultKey is the type holds the request context
type MaxResultKey struct{}

// NodeTwinID is the node twin ID to unmarshal json in it
type NodeTwinID struct {
	TwinID uint32 `json:"twinId"`
}

// NodeData is the NodeData to unmarshal json in it
type NodeData struct {
	NodeResult []NodeTwinID `json:"nodes"`
}

// NodeResult is the NodeData  to unmarshal json in it
type NodeResult struct {
	Data NodeData `json:"data"`
}

// CapacityResult is the NodeData capacity results to unmarshal json in it
type CapacityResult struct {
	Total gridtypes.Capacity `json:"total"`
	Used  gridtypes.Capacity `json:"used"`
}

// NodeInfo is node specific info, queried directly from the node
type NodeInfo struct {
	Capacity   CapacityResult `json:"capacity"`
	DMI        dmi.DMI        `json:"dmi"`
	Hypervisor string         `json:"hypervisor"`
}
