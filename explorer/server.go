package explorer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/zos/pkg/rmb"
)

func (a *App) runServer() {
	log.Info().Str("Server started ... listening on", string(a.explorer)).Msg("")

}

func (a *App) listFarms(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Something bad happened!"))
		return
	}
	log.Debug().Str("request_body", string(body)).Msg("request from external agent")
	queryString := `
	{
		farms {
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
	`
	farmsData := query(queryString)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Something bad happened!"))
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
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Something bad happened!"))
		return
	}
	log.Debug().Str("request_body", string(body)).Msg("request from external agent")
	queryString := fmt.Sprintf(`
	{
		nodes(limit:10%s){
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
		}
		publicConfigs{
			gw4
			ipv4
			ipv6
			gw6
		  }
	}
	`, isSpecificFarm)

	farmsData := query(queryString)

	jsonBytes := []byte(farmsData)

	w.Write(jsonBytes)
}

func (a *App) getNode(w http.ResponseWriter, r *http.Request) {
	nodeId := mux.Vars(r)["node_id"]

	value, err := a.GetRedisKey(fmt.Sprintf("GRID3NODE:%s", nodeId))

	if err != nil {
		log.Warn().Str("Couldn't find entry to redis", string(err.Error())).Msg("")

	}
	if value != "" {
		w.Write([]byte(value))
		return
	}
	twinId := getNodeTwinId(nodeId)
	if twinId < 1 {
		value, err := json.Marshal("Couldn't find node ID.")
		if err != nil {
			log.Error().Err(errors.Wrap(err, "Couldn't get node twin ID")).Msg("")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Something bad happened!"))
			return
		}
		w.Write(value)
		return
	}

	nodeClient := NewNodeClient(twinId, a.rmb)
	nodeCapacity, err := nodeClient.NodeStatistics(a.ctx)
	if err != nil {
		log.Error().Err(errors.Wrap(err, "Couldn't get node statistics")).Msg("connection error")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Something bad happened!"))
		return
	}

	totalCapacity, err := json.Marshal(nodeCapacity)
	if err != nil {
		log.Error().Err(errors.Wrap(err, "Couldn't get node statistics")).Msg("connection error")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Something bad happened!"))
		return
	}

	w.Write(totalCapacity)

	// caching for 30 mins
	err = a.SetRedisKey(fmt.Sprintf("GRID3NODE:%s", nodeId), totalCapacity, 30*60)
	if err != nil {
		log.Error().Err(errors.Wrap(err, "Couldn't cache to redis")).Msg("connection error")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Something bad happened!"))
		return
	}

}
func Setup(router *mux.Router, debug bool, explorer string, redisServer string) {
	log.Info().Msg("Preparing Redis Pool ...")

	redis := &redis.Pool{
		MaxIdle:   10,
		MaxActive: 10,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", "localhost:6379")
			if err != nil {
				log.Error().Err(errors.Wrap(err, "ERROR: fail init redis")).Msg("")
			}
			return conn, err
		},
	}

	rmbClient, err := rmb.Default()
	if err != nil {
		log.Error().Err(errors.Wrap(err, "Couldn't connect to rmb")).Msg("connection error")
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
