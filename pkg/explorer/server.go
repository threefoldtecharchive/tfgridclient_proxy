package explorer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/zos/pkg/gridtypes"
	"github.com/threefoldtech/zos/pkg/rmb"
)

type App struct {
	debug    bool
	explorer string
	redis    *redis.Client
	// what is ctx
	ctx context.Context
}

const (
	Diy       = "DIY"
	Certified = "CERTIFIED"
)

type PublicIP struct {
	ip          int   `json:"ip"`
	gateway     int   `json:"gateway"`
	contract_id int64 `json:"contract_id"`
}

type Farm struct {
	version            int        `json:"version"`
	id                 int        `json:"id"`
	name               []int      `json:"name"`
	twin_id            []int      `json:"twin_id"`
	pricing_policy_id  []int      `json:"pricing_policy_id"`
	certification_type string     `json:"certification_type"`
	country_id         []int      `json:"country_id"`
	city_id            []int      `json:"city_id"`
	public_ips         []PublicIP `json:"public_ips"`
}

type Location struct {
	longitude []int `json:"longitude"`
	latitude  []int `json:"latitude"`
}

type Resources struct {
	hru int64 `json:"hru"`
	sru int64 `json:"sru"`
	cru int64 `json:"cru"`
	mru int64 `json:"mru"`
}

type PublicConfig struct {
	ipv4 []int `json: "ipv4"`
	ipv6 []int `json: "ipv6"`
	gw4  []int `json: "gw4"`
	gw6  []int `json: "gw6"`
}

type Node struct {
	Version           int            `json:"version"`
	Id                int            `json:"id"`
	farm_id           string         `json:"farm_id"`
	twin_id           int64          `json:"twin_id"`
	resources         []Resources    `json:"try"`
	location          []Location     `json:"dat"`
	country_id        int            `json:"country_id"`
	city_id           int            `json:"city_id"`
	public_config     []PublicConfig `json:"public_config"`
	uptime            int64          `json:"uptime"`
	created           int64          `json:"created"`
	farming_policy_id int            `json:"farming_policy_id"`
}

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

func errorReply(message string) []byte {
	return []byte(fmt.Sprintf("{\"status\": \"error\", \"message\": \"%s\"}", message))
}

func NewNodeClient(nodeTwin uint32, bus rmb.Client) *NodeClient {
	return &NodeClient{nodeTwin, bus}
}

func getNodeTwinId(nodeId string) uint32 {

	queryString := fmt.Sprintf(`
	{
		nodes(limit:10, where:{nodeId_eq:%s}){
		  twinId
		}
	}
	`, nodeId)

	result := []byte(query(queryString))

	var res NodeResult
	err := json.Unmarshal(result, &res)

	if err != nil {
		fmt.Println(err)
	}
	return res.Data.NodeResult[0].TwinId

}

func (n *NodeClient) Counters(ctx context.Context) (total CapacityResult, err error) {
	const cmd = "zos.statistics.get"
	var result struct {
		Total gridtypes.Capacity `json:"total"`
		Used  gridtypes.Capacity `json:"used"`
	}
	if err = n.bus.Call(ctx, n.nodeTwin, cmd, nil, &result); err != nil {
		return
	}

	return result, nil
}

func query(jsonQuery string) string {
	jsonData := map[string]string{
		"query": jsonQuery,
	}
	jsonValue, _ := json.Marshal(jsonData)
	request, err := http.NewRequest("POST", "https://explorer.devnet.grid.tf/graphql/", bytes.NewBuffer(jsonValue))
	request.Header.Set("Content-Type", "application/json")
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to graphql network due to %s", err))
	}

	client := &http.Client{Timeout: time.Second * 10}
	response, err := client.Do(request)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to graphql network due to %s", err))
	}
	defer response.Body.Close()
	if err != nil {
		panic(fmt.Sprintf("The HTTP request failed %s", err))
	}
	data, _ := ioutil.ReadAll(response.Body)
	return string(data)
}

func (a *App) run_server() {
	log.Info().Str("Server started ... listening on", string(a.explorer)).Msg("")

	// TODO
	// do queries
	// do cash
}

func (a *App) listFarms(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write(errorReply("couldn't read body"))
		return
	}
	log.Debug().Str("request_body", string(body)).Msg("request from external agent")
	queryString := `
	{
		farms{
		  name
		  farmId
		}
	  }
	`
	farmsData := query(queryString)
	if err != nil {
		err = errors.Wrap(err, "couldn't push entry to reply queue")
		w.Write(errorReply(err.Error()))
	}
	jsonBytes := []byte(farmsData)

	w.Write(jsonBytes)
}

func (a *App) listNodes(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)

	farmId := r.URL.Query().Get("farm_id")
	isSpecificFarm := ""
	if farmId != "" {
		isSpecificFarm = fmt.Sprintf(",where:{farmId_eq:%s}", farmId)
	} else {
		isSpecificFarm = ""
	}

	if err != nil {
		w.Write(errorReply("couldn't read body"))
		return
	}
	log.Debug().Str("request_body", string(body)).Msg("request from external agent")
	queryString := fmt.Sprintf(`
	{
		nodes(limit:10%s){
		  nodeId,
		  twinId
		  publicConfigId
		}
	}
	`, isSpecificFarm)

	farmsData := query(queryString)
	if err != nil {
		err = errors.Wrap(err, "couldn't push entry to reply queue")
		w.Write(errorReply(err.Error()))
	}
	jsonBytes := []byte(farmsData)

	w.Write(jsonBytes)
}

func (a *App) getNode(w http.ResponseWriter, r *http.Request) {
	nodeId := mux.Vars(r)["node_id"]
	twinId := getNodeTwinId(nodeId)
	rmbClient, err := rmb.Default()
	if err != nil {
		panic(err)
	}

	nodeClient := NewNodeClient(twinId, rmbClient)
	nodeCapacity, err := nodeClient.Counters(a.ctx)
	if err != nil {
		panic(err)
	}

	totalCapacity, err := json.Marshal(nodeCapacity)
	if err != nil {
		panic(err)
	}

	w.Write(totalCapacity)

}
func Setup(router *mux.Router, debug bool, explorer string, redisServer string) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisServer,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	a := App{
		debug:    debug,
		explorer: explorer,
		redis:    rdb,
		ctx:      context.Background(),
	}
	go a.run_server()
	router.HandleFunc("/farms", a.listFarms)
	router.HandleFunc("/nodes", a.listNodes)
	router.HandleFunc("/nodes/{node_id:[0-9]+}", a.getNode)
	http.Handle("/", router)
}
