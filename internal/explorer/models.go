package explorer

import (
	"github.com/gomodule/redigo/redis"
	"github.com/patrickmn/go-cache"
	"github.com/threefoldtech/zos/pkg/capacity/dmi"
	"github.com/threefoldtech/zos/pkg/gridtypes"
	"github.com/threefoldtech/zos/pkg/rmb"
)

// DefaultExplorerURL is the default explorer graphql url
const DefaultExplorerURL string = "https://tfchain.dev.threefold.io/graphql/"

// App is the main app objects
type App struct {
	explorer string
	redis    *redis.Pool
	rmb      rmb.Client
	lruCache *cache.Cache
}

// OffsetKey is the type holds the request context
type offsetKey struct{}

// SpecificFarmKey is the type holds the request context
type specificFarmKey struct{}

// MaxResultKey is the type holds the request context
type maxResultKey struct{}

// NodeTwinID is the node twin ID to unmarshal json in it
type nodeTwinID struct {
	TwinID uint32 `json:"twinId"`
}

// NodeData is the NodeData to unmarshal json in it
type nodeData struct {
	NodeResult []nodeTwinID `json:"nodes"`
}

// NodeResult is the NodeData  to unmarshal json in it
type nodeResult struct {
	Data nodeData `json:"data"`
}

// CapacityResult is the NodeData capacity results to unmarshal json in it
type capacityResult struct {
	Total gridtypes.Capacity `json:"total"`
	Used  gridtypes.Capacity `json:"used"`
}

// NodeInfo is node specific info, queried directly from the node
type NodeInfo struct {
	Capacity   capacityResult `json:"capacity"`
	DMI        dmi.DMI        `json:"dmi"`
	Hypervisor string         `json:"hypervisor"`
}
