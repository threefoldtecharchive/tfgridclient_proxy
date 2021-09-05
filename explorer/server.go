package explorer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/zos/client"
	"github.com/threefoldtech/zos/pkg/rmb"
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

	_, err = queryProxy(queryString, w)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
	}
}

func (a *App) listNodes(w http.ResponseWriter, r *http.Request) {
	r, err := a.handleRequestsQueryParams(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
	}

	maxResult := getMaxResult(r.Context())
	pageOffset := getOffset(r.Context())
	isSpecificFarm := getSpecificFarm(r.Context())

	queryString := fmt.Sprintf(`
	{
		nodes(limit:%d,offset:%d,%s){
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
			cru
			mru
			sru
			hru
		publicConfig{
			gw4
			ipv4
			ipv6
			gw6
		  }
		}
	}
	`, maxResult, pageOffset, isSpecificFarm)

	_, err = queryProxy(queryString, w)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
	}
}

func (a *App) getNode(w http.ResponseWriter, r *http.Request) {

	nodeID := mux.Vars(r)["node_id"]
	value, _ := a.GetRedisKey(fmt.Sprintf("GRID3NODE:%s", nodeID))

	// No value, fetch data from the node
	if value == "" {
		nodeInfo, err := a.fetchNodeData(r.Context(), nodeID)
		if err != nil {
			log.Error().Err(err).Msg("could not fetch node data")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(http.StatusText(http.StatusBadRequest)))
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

func (a *App) fetchNodeData(ctx context.Context, nodeID string) (NodeInfo, error) {
	twinID, err := getNodeTwinID(nodeID)
	if err != nil {
		return NodeInfo{}, errors.Wrap(err, "could not get node twin ID")

	}

	nodeClient := client.NewNodeClient(twinID, a.rmb)
	NodeCapacity, UsedCapacity, err := nodeClient.Counters(a.ctx)
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

func (a *App) runServer() {
	log.Info().Str("listening on", a.explorer).Msg("Server started ...")
}

// Setup is the server and do initial configurations
func Setup(router *mux.Router, debug bool, explorer string, redisServer string) {
	log.Info().Msg("Preparing Redis Pool ...")

	redis := &redis.Pool{
		MaxIdle:   10,
		MaxActive: 10,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", "localhost:6379")
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

	a := App{
		debug:    debug,
		explorer: explorer,
		redis:    redis,
		ctx:      context.Background(),
		rmb:      rmbClient,
	}
	go a.runServer()
	router.HandleFunc("/farms", a.listFarms)
	router.HandleFunc("/nodes", a.listNodes)
	router.HandleFunc("/nodes/{node_id:[0-9]+}", a.getNode)
	http.Handle("/", router)
}
