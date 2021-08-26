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

// take the request and restore the query paramas, handle errors and set default values if not available
func (a *App) HandleRequestsQueryParams(r *http.Request) (*http.Request, error) {

	farmID := r.URL.Query().Get("farm_id")
	isSpecificFarm := ""
	if farmID != "" {
		isSpecificFarm = fmt.Sprintf(",where:{farmId_eq:%s}", farmID)
	} else {
		isSpecificFarm = ""
	}

	log.Info().Str("farm", fmt.Sprint(isSpecificFarm)).Msg("Preparing param specific farm id")

	maxResultPerpage := r.URL.Query().Get("max_result")
	if maxResultPerpage == "" {
		maxResultPerpage = "50"
	}

	maxResult, err := strconv.Atoi(maxResultPerpage)
	if err != nil {
		log.Error().Err(errors.Wrap(err, fmt.Sprintf("ERROR: invalid max result number %s", err))).Msg("")
		return &http.Request{}, fmt.Errorf("error: invalid max result number : %w", err)
	}

	log.Info().Str("max result", fmt.Sprint(maxResult)).Msg("Preparing param max result")

	page := r.URL.Query().Get("page")
	if page == "" {
		page = "0"
	}

	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		log.Error().Err(errors.Wrap(err, fmt.Sprintf("ERROR: invalid page number %s", err))).Msg("")
		return &http.Request{}, fmt.Errorf("error: invalid page number : %w", err)
	}

	offset := 0
	if pageNumber > 1 {
		offset = pageNumber * maxResult
	}

	log.Info().Str("offset", fmt.Sprint(offset)).Msg("Preparing param page offset")

	r = r.WithContext(context.WithValue(r.Context(), ContextKey("specific_farm"), isSpecificFarm))
	r = r.WithContext(context.WithValue(r.Context(), ContextKey("page_offset"), offset))
	r = r.WithContext(context.WithValue(r.Context(), ContextKey("max_result"), maxResult))
	return r, nil
}

func (a *App) listFarms(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Something bad happened!"))
		return
	}
	log.Debug().Str("request body", string(body)).Msg("request from external agent")

	r, err = a.HandleRequestsQueryParams(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Something bad happened!"))
	}

	maxResult := r.Context().Value(ContextKey("max_result"))
	pageOffset := r.Context().Value(ContextKey("page_offset"))

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

	r, err = a.HandleRequestsQueryParams(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Something bad happened!"))
	}

	maxResult := r.Context().Value(ContextKey("max_result"))
	pageOffset := r.Context().Value(ContextKey("page_offset"))
	isSpecificFarm := r.Context().Value(ContextKey("specific_farm"))

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
