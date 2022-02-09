package explorer

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/threefoldtech/grid_proxy_server/internal/explorer/db"
	"github.com/threefoldtech/zos/pkg/rmb"
)

// listFarms godoc
// @Summary Show farms on the grid
// @Description Get all farms on the grid, It has pagination
// @Tags GridProxy
// @Accept  json
// @Produce  json
// @Param page query int false "Page number"
// @Param size query int false "Max result per page"
// @Success 200 {object} FarmResult
// @Router /farms [get]
func (a *App) listFarms(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	filter, limit, err := a.handleFarmRequestsQueryParams(r)
	if err != nil {
		errorReplyWithStatus(err, w, http.StatusBadRequest)
		return
	}
	farms, err := a.db.GetFarms(filter, limit)
	if err != nil {
		log.Error().Err(err).Msg("failed to query farm")
		errorReplyWithStatus(err, w, http.StatusInternalServerError)
		return
	}
	serialzied, err := json.Marshal(farms)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal farm")
		errorReplyWithStatus(err, w, http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(serialzied)
}

// getStats godoc
// @Summary Show stats about the grid
// @Description Get statistics about the grid
// @Tags GridProxy
// @Accept  json
// @Produce  json
// @Success 200 {object} FarmResult
// @Router /stats [get]
func (a *App) getStats(w http.ResponseWriter, r *http.Request) {
	counters, err := a.db.GetCounters()
	if err != nil {
		errorReplyWithStatus(err, w, http.StatusInternalServerError)
		return
	}
	serialized, err := json.Marshal(counters)
	if err != nil {
		errorReplyWithStatus(err, w, http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(serialized)
}

// listNodes godoc
// @Summary Show nodes on the grid
// @Description Get all nodes on the grid, It has pagination
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
	filter, limit, err := a.handleNodeRequestsQueryParams(r)
	if err != nil {
		errorReplyWithStatus(err, w, http.StatusBadRequest)
		return
	}
	dbNodes, err := a.db.GetNodes(filter, limit)
	if err != nil {
		errorReplyWithStatus(err, w, http.StatusInternalServerError)
		return
	}
	nodes := make([]node, len(dbNodes))
	for idx, node := range dbNodes {
		nodes[idx] = nodeFromDBNode(node)

	}

	// return the number of pages and totalCount in the response headers
	nodesCount, err := a.getTotalCount()
	if err != nil {
		log.Error().Err(err).Msg("error fetching pages")
	} else {
		pages := math.Ceil(float64(nodesCount) / float64(limit.Size))
		w.Header().Add("count", fmt.Sprintf("%d", nodesCount))
		w.Header().Add("size", fmt.Sprintf("%d", limit.Size))
		w.Header().Add("pages", fmt.Sprintf("%d", int(pages)))
	}

	result, err := json.Marshal(nodes)
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
	nodeData, err := a.getNodeData(nodeID)
	if err != nil {
		errorReply(err, w)
		return
	}
	serialized, err := json.Marshal(nodeData)
	if err != nil {
		errorReply(err, w)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(serialized)
}

func (a *App) getNodeStatus(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	w.Header().Add("Content-Type", "application/json")

	response := NodeStatus{}
	nodeID := mux.Vars(r)["node_id"]

	nodeData, err := a.getNodeData(nodeID)
	if err != nil {
		errorReply(err, w)
		return
	}
	response.Status = nodeData.Status
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
func Setup(router *mux.Router, redisServer string, gitCommit string, postgresHost string, postgresPort int, postgresDB, postgresUser, postgresPassword string) error {
	log.Info().Str("redis address", redisServer).Msg("Preparing Redis Pool ...")

	rmbClient, err := rmb.NewClient("tcp://127.0.0.1:6379", 500)
	if err != nil {
		return errors.Wrap(err, "couldn't connect to rmb")
	}
	c := cache.New(2*time.Minute, 3*time.Minute)
	db, err := db.NewPostgresDatabase(postgresHost, postgresPort, postgresUser, postgresPassword, postgresDB)
	if err != nil {
		return errors.Wrap(err, "couldn't get sqlite3 client")
	}
	a := App{
		db:             db,
		rmb:            rmbClient,
		lruCache:       c,
		releaseVersion: gitCommit,
	}

	router.HandleFunc("/farms", a.listFarms)
	router.HandleFunc("/stats", a.getStats)
	router.HandleFunc("/nodes", a.listNodes)
	router.HandleFunc("/gateways", a.listNodes)
	router.HandleFunc("/nodes/{node_id:[0-9]+}", a.getNode)
	router.HandleFunc("/gateways/{node_id:[0-9]+}", a.getNode)
	router.HandleFunc("/nodes/{node_id:[0-9]+}/status", a.getNodeStatus)
	router.HandleFunc("/gateways/{node_id:[0-9]+}/status", a.getNodeStatus)
	router.HandleFunc("/", a.indexPage)
	router.HandleFunc("/version", a.version)
	router.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)
	nodeManager := NewNodeManager(db, rmbClient, 30)
	go nodeManager.Run(context.Background())
	return nil
}
