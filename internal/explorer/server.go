package explorer

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/threefoldtech/zos/pkg/rmb"
)

// listFarms godoc
// @Summary Show farms on the grid
// @Description Get all farms on the grid from graphql, It has pagination
// @Tags GridProxy
// @Accept  json
// @Produce  json
// @Param page query int false "Page number"
// @Param size query int false "Max result per page"
// @Success 200 {object} FarmResult
// @Router /farms [get]
func (a *App) listFarms(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	r, err := a.handleRequestsQueryParams(r)
	if err != nil {
		errorReplyWithStatus(err, w, http.StatusBadRequest)
		return
	}
	maxResult, pageOffset := getMaxResult(r.Context()), getOffset(r.Context())

	queryString := fmt.Sprintf(`
	{
		farms (limit:%d,offset:%d) {
			name
			farmId
			twinId
			version
			farmId
			pricingPolicyId
			stellarAddress
			publicIPs{
				id
				ip
				contractId
				gateway
			}
		}
	}
	`, maxResult, pageOffset)

	farms := FarmResult{}
	err = a.query(queryString, &farms)

	if err != nil {
		log.Error().Err(err).Msg("failed to query farm")
		errorReplyWithStatus(err, w, http.StatusInternalServerError)
		return
	}

	result, err := json.Marshal(farms.Data.Farms)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal farm")
		errorReplyWithStatus(err, w, http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

// listNodes godoc
// @Summary Show nodes on the grid
// @Description Get all nodes on the grid from graphql, It has pagination
// @Tags GridProxy
// @Accept  json
// @Produce  json
// @Param page query int false "Page number"
// @Param size query int false "Max result per page"
// @Param farm_id query int false "Get nodes for specific farm"
// @Success 200 {object} nodesResponse
// @Router /nodes [get]
// @Router /gateways [get]
func (a *App) listNodes(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	r, err := a.handleRequestsQueryParams(r)
	if err != nil {
		errorReplyWithStatus(err, w, http.StatusBadRequest)
		return
	}

	maxResult := getMaxResult(r.Context())
	pageOffset := getOffset(r.Context())
	isSpecificFarm := getSpecificFarm(r.Context())
	isGateway := getIsGateway(r.Context())

	nodes, err := a.getAllNodes(maxResult, pageOffset, isSpecificFarm, isGateway)

	if err != nil {
		log.Error().Err(err).Msg("fail to list nodes")
		errorReplyWithStatus(err, w, http.StatusInternalServerError)
		return
	}
	var nodeList []node
	for _, node := range nodes.Nodes.Data {
		// check if node is down or not by checking redis key existence in redis cache and if it is down
		// set status to down in node struct and add it to nodeList slice
		nodeInfoStr, err := a.GetRedisKey(a.getNodeKey(fmt.Sprint(node.NodeID)))
		if err != nil {
			node.Status = "down"
		}
		if nodeInfoStr == "" {
			node.Status = "down"
		} else {
			var nodeInfo NodeInfo
			err := json.Unmarshal([]byte(nodeInfoStr), &nodeInfo)
			if err != nil {
				log.Error().Err(err).Msg("failed to unmarshal redis data to struct")
			}
			node.Status = "up"
			node.TotalResources = nodeInfo.Capacity.Total
			node.UsedResources = nodeInfo.Capacity.Used

		}

		node.Location.City = node.City
		node.Location.Country = node.Country

		nodeList = append(nodeList, node)
	}

	// return the number of pages and totalCount in the response headers
	nodesCount, err := a.getTotalCount()
	if err != nil {
		log.Error().Err(err).Msg("error fetching pages")
	} else {
		pages := math.Ceil(float64(nodesCount) / float64(maxResult)) 
		w.Header().Add("count", fmt.Sprintf("%d", nodesCount))
		w.Header().Add("size", fmt.Sprintf("%d", maxResult))
		w.Header().Add("pages", fmt.Sprintf("%d", int(pages)))
	}

	result, err := json.Marshal(nodeList)
	if err != nil {
		log.Error().Err(err).Msg("fail to list nodes")
		errorReply(err, w)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

// getNode godoc
// @Summary Show the details for specific node
// @Description Get all details for specific node hardware, capacity, DMI, hypervisor
// @Tags GridProxy
// @Param node_id path int false "Node ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} NodeInfo
// @Router /nodes/{node_id} [get]
// @Router /gateways/{node_id} [get]
func (a *App) getNode(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	w.Header().Add("Content-Type", "application/json")

	nodeID := mux.Vars(r)["node_id"]
	nodeData, err := a.getNodeData(nodeID, false)
	if err != nil {
		errorReply(err, w)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(nodeData))
}

func (a *App) getNodeStatus(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	w.Header().Add("Content-Type", "application/json")

	response := NodeStatus{}
	nodeID := mux.Vars(r)["node_id"]

	nodeInfo, err := a.GetRedisKey(a.getNodeKey(fmt.Sprint(nodeID)))
	if err != nil {
		errorReply(err, w)
		return
	} else if nodeInfo == "" {
		response.Status = "down"
	} else {
		response.Status = "up"
	}
	w.WriteHeader(http.StatusOK)
	res, _ := response.Serialize()
	w.Write(res)
}

func (a *App) indexPage(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("welcome to grid proxy server, available endpoints [/farms, /nodes, /nodes/<node-id>]"))
}

func (a *App) version(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("{\"version\": \"%s\"}", a.releaseVersion)))
}

// Setup is the server and do initial configurations
// @title Grid Proxy Server API
// @version 1.0
// @description grid proxy server has the main methods to list farms, nodes, node details in the grid.
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /
func Setup(router *mux.Router, explorer string, redisServer string, gitCommit string) {
	log.Info().Str("redis address", redisServer).Msg("Preparing Redis Pool ...")

	redis := &redis.Pool{
		MaxIdle:   20,
		MaxActive: 100,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", redisServer)
			if err != nil {
				log.Error().Err(err).Msg("fail init redis")
			}
			return conn, err
		},
	}

	rmbClient, err := rmb.NewClient("tcp://127.0.0.1:6379", 500)
	if err != nil {
		log.Error().Err(err).Msg("couldn't connect to rmb")
		return
	}
	c := cache.New(2*time.Minute, 3*time.Minute)

	a := App{
		explorer:       explorer,
		redis:          redis,
		rmb:            rmbClient,
		lruCache:       c,
		releaseVersion: gitCommit,
	}

	router.HandleFunc("/farms", a.listFarms)
	router.HandleFunc("/nodes", a.listNodes)
	router.HandleFunc("/gateways", a.listNodes)
	router.HandleFunc("/nodes/{node_id:[0-9]+}", a.getNode)
	router.HandleFunc("/gateways/{node_id:[0-9]+}", a.getNode)
	router.HandleFunc("/nodes/{node_id:[0-9]+}/status", a.getNodeStatus)
	router.HandleFunc("/gateways/{node_id:[0-9]+}/status", a.getNodeStatus)
	router.HandleFunc("/", a.indexPage)
	router.HandleFunc("/version", a.version)
	router.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)
	// Run node caching every 2 minutes
	go a.cacheNodesInfo()
	job := cron.New()
	job.AddFunc("@every 2m", a.cacheNodesInfo)
	job.Start()
}
