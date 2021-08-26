package explorer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

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
	_, err = queryProxy(queryString, w)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Something bad happened!"))
	}

}

func (a *App) listNodes(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Something bad happened!"))
		return
	}
	log.Debug().Str("request_body", string(body)).Msg("request from external agent")

	farmID := r.URL.Query().Get("farm_id")
	isSpecificFarm := ""
	if farmID != "" {
		isSpecificFarm = fmt.Sprintf(",where:{farmId_eq:%s}", farmID)
	} else {
		isSpecificFarm = ""
	}
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "0"
	}
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 - Bad page number!"))
		return
	}
	offset := 0
	if pageNumber > 1 {
		offset = pageNumber * 50
	}
	log.Debug().Str("page number", fmt.Sprint(pageNumber)).Msg("Requested page number")
	queryString := fmt.Sprintf(`
	{
		nodes(limit:50,offset:%d,%s){
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
	`, offset, isSpecificFarm)

	_, err = queryProxy(queryString, w)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Something bad happened!"))
	}
}

func (a *App) getNode(w http.ResponseWriter, r *http.Request) {
	nodeID := mux.Vars(r)["node_id"]

	value, err := a.GetRedisKey(fmt.Sprintf("GRID3NODE:%s", nodeID))

	if err != nil {
		log.Warn().Str("Couldn't find entry to redis", string(err.Error())).Msg("")

	}
	if value != "" {
		w.Write([]byte(value))
		return
	}
	TwinID, err := getNodeTwinID(nodeID)
	if err != nil {

		log.Error().Err(errors.Wrap(err, "Couldn't get node twin ID")).Msg("")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Something bad happened!"))
		return

	}

	nodeClient := NewNodeClient(TwinID, a.rmb)
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

	// caching for 30 mins
	err = a.SetRedisKey(fmt.Sprintf("GRID3NODE:%s", nodeID), totalCapacity, 30*60)
	if err != nil {
		log.Fatal().Err(errors.Wrap(err, "Couldn't cache to redis")).Msg("connection error")
	}

	w.Write(totalCapacity)

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
