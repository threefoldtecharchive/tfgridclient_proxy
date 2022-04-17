package explorer

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	httpSwagger "github.com/swaggo/http-swagger"

	// swagger configuration
	_ "github.com/threefoldtech/grid_proxy_server/docs"
	"github.com/threefoldtech/grid_proxy_server/internal/explorer/db"
	"github.com/threefoldtech/grid_proxy_server/internal/explorer/mw"
	"github.com/threefoldtech/zos/pkg/rmb"
)

const (
	// SSDOverProvisionFactor factor by which the ssd are allowed to be overprovisioned
	SSDOverProvisionFactor = 2
)

// listFarms godoc
// @Summary Show farms on the grid
// @Description Get all farms on the grid, It has pagination
// @Tags GridProxy
// @Accept  json
// @Produce  json
// @Param page query int false "Page number"
// @Param size query int false "Max result per page"
// @Param ret_count query string false "Set farms' count on headers based on filter"
// @Param free_ips query int false "Min number of free ips in the farm"
// @Param pricing_policy_id query int false "Pricing policy id"
// @Param version query int false "farm version"
// @Param farm_id query int false "farm id"
// @Param twin_id query int false "twin id associated with the farm"
// @Param name query string false "farm name"
// @Param stellar_address query string false "farm stellar_address"
// @Success 200 {object} []db.Farm
// @Router /farms [get]
func (a *App) listFarms(r *http.Request) (interface{}, mw.Response) {
	filter, limit, err := a.handleFarmRequestsQueryParams(r)
	if err != nil {
		return nil, mw.BadRequest(err)
	}
	dbFarms, err := a.db.GetFarms(filter, limit)
	if err != nil {
		log.Error().Err(err).Msg("failed to query farm")
		return nil, mw.Error(err)
	}

	var farmsCount uint
	farms := make([]farm, len(dbFarms))
	for idx, farm := range dbFarms {
		farmsCount, farms[idx] = farmFromDBFarm(farm)

	}
	resp := mw.Ok()

	// return the number of pages and totalCount in the response headers
	if limit.RetCount {
		pages := math.Ceil(float64(farmsCount) / float64(limit.Size))
		resp = resp.WithHeader("count", fmt.Sprintf("%d", farmsCount)).
			WithHeader("size", fmt.Sprintf("%d", limit.Size)).
			WithHeader("pages", fmt.Sprintf("%d", int(pages)))
	}
	return farms, resp
}

// getStats godoc
// @Summary Show stats about the grid
// @Description Get statistics about the grid
// @Tags GridProxy
// @Accept  json
// @Produce  json
// @Param status query string false "Node status filter, up/down."
// @Success 200 {object} []db.Counters
// @Router /stats [get]
func (a *App) getStats(r *http.Request) (interface{}, mw.Response) {
	filter, err := a.handleStatsRequestsQueryParams(r)
	counters, err := a.db.GetCounters(filter)
	if err != nil {
		return nil, mw.Error(err)
	}
	return counters, nil
}

// listNodes godoc
// @Summary Show nodes on the grid
// @Description Get all nodes on the grid, It has pagination
// @Tags GridProxy
// @Accept  json
// @Produce  json
// @Param page query int false "Page number"
// @Param size query int false "Max result per page"
// @Param ret_count query string false "Set nodes' count on headers based on filter"
// @Param free_mru query int false "Min free reservable mru in bytes"
// @Param free_hru query int false "Min free reservable hru in bytes"
// @Param free_sru query int false "Min free reservable sru in bytes"
// @Param free_ips query int false "Min number of free ips in the farm of the node"
// @Param status query string false "Node status filter, up/down."
// @Param city query string false "Node city filter"
// @Param country query string false "Node country filter"
// @Param farm_name query string false "Get nodes for specific farm"
// @Param ipv4 query string false "Set to true to filter nodes with ipv4"
// @Param ipv6 query string false "Set to true to filter nodes with ipv6"
// @Param domain query string false "Set to true to filter nodes with domain"
// @Param farm_ids query string false "List of farms separated by comma to fetch nodes from (e.g. '1,2,3')"
// @Success 200 {object} []node
// @Router /nodes [get]
// @Router /gateways [get]
func (a *App) listNodes(r *http.Request) (interface{}, mw.Response) {
	filter, limit, err := a.handleNodeRequestsQueryParams(r)
	if err != nil {
		return nil, mw.BadRequest(err)
	}
	dbNodes, err := a.db.GetNodes(filter, limit)
	if err != nil {
		return nil, mw.Error(err)
	}
	var nodesCount uint
	nodes := make([]node, len(dbNodes))
	for idx, node := range dbNodes {
		nodesCount, nodes[idx] = nodeFromDBNode(node)

	}
	resp := mw.Ok()

	// return the number of pages and totalCount in the response headers
	if !limit.RetCount {
		nodesCount, err = a.getTotalCount()
		if err != nil {
			log.Error().Err(err).Msg("error fetching pages")
		}
	}
	pages := math.Ceil(float64(nodesCount) / float64(limit.Size))
	resp = resp.WithHeader("count", fmt.Sprintf("%d", nodesCount)).
		WithHeader("size", fmt.Sprintf("%d", limit.Size)).
		WithHeader("pages", fmt.Sprintf("%d", int(pages)))

	return nodes, resp
}

// getNode godoc
// @Summary Show the details for specific node
// @Description Get all details for specific node hardware, capacity, DMI, hypervisor
// @Tags GridProxy
// @Param node_id path int false "Node ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} node
// @Router /nodes/{node_id} [get]
// @Router /gateways/{node_id} [get]
func (a *App) getNode(r *http.Request) (interface{}, mw.Response) {
	nodeID := mux.Vars(r)["node_id"]
	nodeData, err := a.getNodeData(nodeID)
	if err != nil {
		return nil, errorReply(err)
	}
	return nodeData, nil
}

func (a *App) getNodeStatus(r *http.Request) (interface{}, mw.Response) {
	response := NodeStatus{}
	nodeID := mux.Vars(r)["node_id"]

	nodeData, err := a.getNodeData(nodeID)
	if err != nil {
		return nil, errorReply(err)
	}
	response.Status = nodeData.Status
	return response, nil
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
// @BasePath /
func Setup(router *mux.Router, redisServer string, gitCommit string, database db.Database) error {
	log.Info().Str("redis address", redisServer).Msg("Preparing Redis Pool ...")

	rmbClient, err := rmb.NewClient(redisServer, 500)
	if err != nil {
		return errors.Wrap(err, "couldn't connect to rmb")
	}
	c := cache.New(2*time.Minute, 3*time.Minute)
	a := App{
		db:             database,
		rmb:            rmbClient,
		lruCache:       c,
		releaseVersion: gitCommit,
	}

	router.HandleFunc("/farms", mw.AsHandlerFunc(a.listFarms))
	router.HandleFunc("/stats", mw.AsHandlerFunc(a.getStats))
	router.HandleFunc("/nodes", mw.AsHandlerFunc(a.listNodes))
	router.HandleFunc("/gateways", mw.AsHandlerFunc(a.listNodes))
	router.HandleFunc("/nodes/{node_id:[0-9]+}", mw.AsHandlerFunc(a.getNode))
	router.HandleFunc("/gateways/{node_id:[0-9]+}", mw.AsHandlerFunc(a.getNode))
	router.HandleFunc("/nodes/{node_id:[0-9]+}/status", mw.AsHandlerFunc(a.getNodeStatus))
	router.HandleFunc("/gateways/{node_id:[0-9]+}/status", mw.AsHandlerFunc(a.getNodeStatus))
	router.HandleFunc("/", a.indexPage)
	router.HandleFunc("/version", a.version)
	router.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)
	return nil
}
