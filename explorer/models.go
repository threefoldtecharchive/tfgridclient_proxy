package explorer

import (
	"context"

	"github.com/gomodule/redigo/redis"
	"github.com/threefoldtech/zos/pkg/gridtypes"
	"github.com/threefoldtech/zos/pkg/rmb"
)

// App is the main app objects
type App struct {
	debug    bool
	explorer string
	redis    *redis.Pool
	ctx      context.Context
	rmb      rmb.Client
}

// const (
// 	Diy       = "DIY"
// 	Certified = "CERTIFIED"
// )

// type PublicIP struct {
// 	ip          int   `json:"ip"`
// 	gateway     int   `json:"gateway"`
// 	contract_id int64 `json:"contract_id"`
// }

// type Farm struct {
// 	version            int        `json:"version"`
// 	id                 int        `json:"id"`
// 	name               []int      `json:"name"`
// 	twin_id            []int      `json:"twin_id"`
// 	pricing_policy_id  []int      `json:"pricing_policy_id"`
// 	certification_type string     `json:"certification_type"`
// 	country_id         []int      `json:"country_id"`
// 	city_id            []int      `json:"city_id"`
// 	public_ips         []PublicIP `json:"public_ips"`
// }

// type Location struct {
// 	longitude []int `json:"longitude"`
// 	latitude  []int `json:"latitude"`
// }

// type Resources struct {
// 	hru int64 `json:"hru"`
// 	sru int64 `json:"sru"`
// 	cru int64 `json:"cru"`
// 	mru int64 `json:"mru"`
// }

// type PublicConfig struct {
// 	ipv4 []int `json: "ipv4"`
// 	ipv6 []int `json: "ipv6"`
// 	gw4  []int `json: "gw4"`
// 	gw6  []int `json: "gw6"`
// }

// type Node struct {
// 	Version           int            `json:"version"`
// 	Id                int            `json:"id"`
// 	farm_id           string         `json:"farm_id"`
// 	twin_id           int64          `json:"twin_id"`
// 	resources         []Resources    `json:"try"`
// 	location          []Location     `json:"dat"`
// 	country_id        int            `json:"country_id"`
// 	city_id           int            `json:"city_id"`
// 	public_config     []PublicConfig `json:"public_config"`
// 	uptime            int64          `json:"uptime"`
// 	created           int64          `json:"created"`
// 	farming_policy_id int            `json:"farming_policy_id"`
// }

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

// NodeClient is the Nodeclient  to unmarshal json in it
type NodeClient struct {
	nodeTwin uint32
	bus      rmb.Client
}

// CapacityResult is the NodeData capacity results to unmarshal json in it
type CapacityResult struct {
	Total gridtypes.Capacity `json:"total"`
	Used  gridtypes.Capacity `json:"used"`
}
