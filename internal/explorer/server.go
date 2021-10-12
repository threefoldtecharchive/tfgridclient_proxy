package explorer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/zos/client"
	"github.com/threefoldtech/zos/pkg/rmb"
)

// ErrNodeNotFound creates new error type to define node existence or server problem
var (
	ErrNodeNotFound = errors.New("node not found")
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
	queryString := fmt.Sprintf(`
	{
		nodes(limit:%d,offset:%d, where:{%s%s}){
			version          
			id
			nodeId        
			farmId          
			twinId          
			country
			gridVersion  
			city         
			uptime           
			created          
			farmingPolicyId
			updatedAt
			cru
			mru
			sru
			hru
		publicConfig{
			domain
			gw4
			gw6
			ipv4
			ipv6
		  }
		}
	}
	`, maxResult, pageOffset, isSpecificFarm, isGateway)

	_, err = a.queryProxy(queryString, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
}

func (a *App) getNode(w http.ResponseWriter, r *http.Request) {

	nodeID := mux.Vars(r)["node_id"]
	value, _ := a.GetRedisKey(fmt.Sprintf("GRID3NODE:%s", nodeID))

	// No value, fetch data from the node
	if value == "" {
		nodeInfo, err := a.fetchNodeData(r.Context(), nodeID)
		if errors.Is(err, ErrNodeNotFound) {
			// return not found 404
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(http.StatusText(http.StatusNotFound)))
			return
		} else if err != nil {
			// return internal server error
			log.Error().Err(err).Msg("could not fetch node data")
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(http.StatusText(http.StatusBadGateway)))
			return
		}

		// Save value in redis
		// caching for 30 mins
		marshalledInfo, err := json.Marshal(nodeInfo)
		if err != nil {
			log.Error().Err(err).Msg("could not marshal node info")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			return
		}
		err = a.SetRedisKey(fmt.Sprintf("GRID3NODE:%s", nodeID), marshalledInfo, 30*60)
		if err != nil {
			log.Warn().Err(err).Msg("could not cache data in redis")
		}
		value = string(marshalledInfo)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	w.Write([]byte(value))
}

func (a *App) indexPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("welcome to grid proxy server, available endpoints [/farms, /nodes, /nodes/<node-id>]"))
}

func (a *App) fetchNodeData(ctx context.Context, nodeID string) (NodeInfo, error) {
	twinID, err := a.getNodeTwinID(nodeID)
	if err != nil {
		return NodeInfo{}, ErrNodeNotFound

	}

	nodeClient := client.NewNodeClient(twinID, a.rmb)
	NodeCapacity, UsedCapacity, err := nodeClient.Counters(ctx)
	if err != nil {
		return NodeInfo{}, errors.Wrap(err, "could not get node capacity")
	}

	dmi, err := nodeClient.SystemDMI(ctx)
	if err != nil {
		return NodeInfo{}, errors.Wrap(err, "could not get node DMI info")
	}

	hypervisor, err := nodeClient.SystemHypervisor(ctx)
	if err != nil {
		return NodeInfo{}, errors.Wrap(err, "could not get node hypervisor info")
	}

	capacity := capacityResult{}
	capacity.Total = NodeCapacity
	capacity.Used = UsedCapacity

	return NodeInfo{
		Capacity:   capacity,
		DMI:        dmi,
		Hypervisor: hypervisor,
	}, nil

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
}
