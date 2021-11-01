package explorer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/zos/pkg/rmb"
)

// ErrNodeNotFound creates new error type to define node existence or server problem
var (
	ErrNodeNotFound = errors.New("node not found")
)

// ErrBadGateway creates new error type to define node existence or server problem
var (
	ErrBadGateway = errors.New("bad gateway")
)

func (a *App) listFarms(w http.ResponseWriter, r *http.Request) {
	r, err := a.handleRequestsQueryParams(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
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
		}
		publicIps{
			id
			ip
			farmId
			contractId
			gateway
			
		}
	}
	`, maxResult, pageOffset)

	_, err = a.queryProxy(queryString, w)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
}

func (a *App) listNodes(w http.ResponseWriter, r *http.Request) {
	r, err := a.handleRequestsQueryParams(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	maxResult := getMaxResult(r.Context())
	pageOffset := getOffset(r.Context())
	isSpecificFarm := getSpecificFarm(r.Context())
	isGateway := getIsGateway(r.Context())

	nodes, err := a.getAllNodes(maxResult, pageOffset, isSpecificFarm, isGateway)

	if err != nil {
		log.Error().Err(err).Msg("fail to list nodes")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	var nodeList []node
	for _, node := range nodes.Nodes.Data {
		isStored, err := a.GetRedisKey(fmt.Sprintf("GRID3NODE:%d", node.NodeID))
		if err != nil {
			node.State = "down"
		}
		if isStored != "" {
			node.State = "up"
		}
		nodeList = append(nodeList, node)
	}

	result, err := json.Marshal(nodeList)
	if err != nil {
		log.Error().Err(err).Msg("fail to list nodes")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

func (a *App) getNode(w http.ResponseWriter, r *http.Request) {

	nodeID := mux.Vars(r)["node_id"]
	nodeData, err := a.getNodeData(nodeID)
	if errors.Is(err, ErrNodeNotFound) {
		// return not found 404
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	} else if errors.Is(err, ErrBadGateway) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(http.StatusText(http.StatusBadGateway)))
		return
	} else if err != nil {
		// return internal server error
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(nodeData))
}

func (a *App) indexPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("welcome to grid proxy server, available endpoints [/farms, /nodes, /nodes/<node-id>]"))
}

// Setup is the server and do initial configurations
func Setup(router *mux.Router, explorer string, redisServer string) {
	log.Info().Str("redis address", redisServer).Msg("Preparing Redis Pool ...")

	redis := &redis.Pool{
		MaxIdle:   10,
		MaxActive: 10,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", redisServer)
			if err != nil {
				log.Error().Err(err).Msg("fail init redis")
			}
			return conn, err
		},
	}

	rmbClient, err := rmb.Default()
	if err != nil {
		log.Error().Err(err).Msg("couldn't connect to rmb")
		return
	}
	c := cache.New(10*time.Minute, 15*time.Minute)

	a := App{
		explorer: explorer,
		redis:    redis,
		rmb:      rmbClient,
		lruCache: c,
	}

	router.HandleFunc("/farms", a.listFarms)
	router.HandleFunc("/nodes", a.listNodes)
	router.HandleFunc("/gateways", a.listNodes)
	router.HandleFunc("/nodes/{node_id:[0-9]+}", a.getNode)
	router.HandleFunc("/gateways/{node_id:[0-9]+}", a.getNode)
	router.HandleFunc("/", a.indexPage)
	// Run node caching every 30 minutes
	go a.cacheNodesInfo()
	job := cron.New()
	job.AddFunc("@every 30m", a.cacheNodesInfo)
	job.Start()
}
