package explorer

import (
	"context"

	"github.com/gomodule/redigo/redis"
	"github.com/threefoldtech/zos/pkg/gridtypes"
	"github.com/threefoldtech/zos/pkg/rmb"
)

type App struct {
	debug    bool
	explorer string
	redis    *redis.Pool
	ctx      context.Context
}

const (
	Diy       = "DIY"
	Certified = "CERTIFIED"
)

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

type NodeTwinId struct {
	TwinId uint32 `json:"twinId"`
}

type NodeData struct {
	NodeResult []NodeTwinId `json:"nodes"`
}
type NodeResult struct {
	Data NodeData `json:"data"`
}

type NodeClient struct {
	nodeTwin uint32
	bus      rmb.Client
}

type CapacityResult struct {
	Total gridtypes.Capacity `json:"total"`
	Used  gridtypes.Capacity `json:"used"`
}
