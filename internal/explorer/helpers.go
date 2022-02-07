package explorer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/zos/client"
)

const maxGoRoutnes = 30 // limit go routines so we have 30 node per time

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func errorReplyWithStatus(err error, w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	var res ErrorReply
	res.Message = err.Error()
	res.Error = http.StatusText(status)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		w.Write([]byte(`{"error": "Internal server error", "message": "couldn't encode json"}`))
	}
}
func errorReply(err error, w http.ResponseWriter) {
	var res ErrorReply
	res.Message = err.Error()
	if errors.Is(err, ErrNodeNotFound) {
		// return not found 404
		w.WriteHeader(http.StatusNotFound)
		res.Error = http.StatusText(http.StatusNotFound)
	} else if errors.Is(err, ErrBadGateway) {
		w.WriteHeader(http.StatusBadGateway)
		res.Error = http.StatusText(http.StatusBadGateway)
	} else if err != nil {
		// return internal server error
		log.Error().Err(err).Msg("failed to get node information")
		w.WriteHeader(http.StatusInternalServerError)
		res.Error = http.StatusText(http.StatusInternalServerError)
	}
	if err := json.NewEncoder(w).Encode(res); err != nil {
		w.Write([]byte(`{"error": "Internal server error", "message": "couldn't encode json"}`))
	}
}

func (a *App) getNodeKey(nodeID string) string {
	return fmt.Sprintf("GRID3NODE:%s", nodeID)
}

func (a *App) getNodeCapacityKey(nodeID string) string {
	return fmt.Sprintf("GRID3NODECAPACITY:%s", nodeID)
}

func (a *App) getNodeTwinID(nodeID string) (uint32, error) {
	// cache node twin id for 10 mins and purge after 15
	if twinID, found := a.lruCache.Get(nodeID); found {
		return twinID.(uint32), nil
	}

	queryString := fmt.Sprintf(`
	{
		nodes(where:{nodeId_eq:%s}){
		  twinId
		}
	}
	`, nodeID)

	var res nodeResult
	err := a.query(queryString, &res)

	if err != nil {
		return 0, fmt.Errorf("failed to query node %w", err)
	}

	nodeStats := res.Data.NodeResult
	if len(nodeStats) < 1 {
		return 0, ErrNodeNotFound
	}

	twinID := nodeStats[0].TwinID
	a.lruCache.Set(nodeID, twinID, cache.DefaultExpiration)
	return twinID, nil
}

func (a *App) baseQuery(queryString string) (io.ReadCloser, error) {
	jsonData := map[string]string{
		"query": queryString,
	}
	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		return nil, fmt.Errorf("invalid query string %w", err)
	}

	request, err := http.NewRequest("POST", a.explorer, bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, fmt.Errorf("failed to query explorer network %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: time.Second * 10}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to query explorer network %w", err)
	}

	if response.StatusCode == 200 {
		return response.Body, nil
	}

	var errResult interface{}
	if err := json.NewDecoder(response.Body).Decode(errResult); err != nil {
		return nil, fmt.Errorf("failed to decode error from page: %w", err)
	}
	return nil, fmt.Errorf("failed to query explorer network: %v", errResult)
}

func (a *App) query(queryString string, result interface{}) error {
	response, err := a.baseQuery(queryString)
	if err != nil {
		return err
	}
	defer response.Close()

	if err := json.NewDecoder(response).Decode(result); err != nil {
		return err
	}

	return nil
}

func getOffset(ctx context.Context) int {
	return ctx.Value(offsetKey{}).(int)
}

func getMaxResult(ctx context.Context) int {
	return ctx.Value(maxResultKey{}).(int)
}

func getSpecificFarm(ctx context.Context) string {
	return ctx.Value(specificFarmKey{}).(string)
}

func getIsGateway(ctx context.Context) string {
	return ctx.Value(isGatewayKey{}).(string)
}

func calculateMaxResult(r *http.Request) (int, error) {
	maxResultPerpage := r.URL.Query().Get("size")
	if maxResultPerpage == "" {
		maxResultPerpage = "50"
	}

	maxResult, err := strconv.Atoi(maxResultPerpage)
	if err != nil {
		return 0, fmt.Errorf("invalid page number : %w", err)
	}

	return maxResult, nil
}

func calculateOffset(maxResult int, r *http.Request) (int, error) {
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}

	pageNumber, err := strconv.Atoi(page)
	if err != nil || pageNumber < 1 {
		return 0, fmt.Errorf("invalid page number : %w", err)
	}
	return (pageNumber - 1) * maxResult, nil
}

// HandleRequestsQueryParams takes the request and restore the query paramas, handle errors and set default values if not available
func (a *App) handleRequestsQueryParams(r *http.Request) (*http.Request, error) {
	isGateway := ""
	if strings.Contains(fmt.Sprint(r.URL), "gateways") {
		isGateway = `,publicConfig_json: {domain_contains:"."}`
	} else {
		isGateway = ""
	}

	farmID := r.URL.Query().Get("farm_id")
	isSpecificFarm := ""
	if farmID != "" {
		isSpecificFarm = fmt.Sprintf(",farmId_eq:%s", farmID)
	} else {
		isSpecificFarm = ""
	}

	maxResult, err := calculateMaxResult(r)
	if err != nil {
		return nil, err
	}
	offset, err := calculateOffset(maxResult, r)
	if err != nil {
		return nil, err
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, specificFarmKey{}, isSpecificFarm)
	ctx = context.WithValue(ctx, offsetKey{}, offset)
	ctx = context.WithValue(ctx, maxResultKey{}, maxResult)
	ctx = context.WithValue(ctx, isGatewayKey{}, isGateway)
	return r.WithContext(ctx), nil
}

func (a *App) getNodeHypervisor(ctx context.Context, nodeID string, nodeClient *client.NodeClient) (string, error) {
	nodeKey := fmt.Sprintf("node_%s_hypervisor", nodeID)
	if nodeHyperVisor, found := a.lruCache.Get(nodeKey); found {
		return nodeHyperVisor.(string), nil
	}

	hypervisor, err := nodeClient.SystemHypervisor(ctx)

	if err != nil {
		return "", err
	}

	a.lruCache.Set(nodeKey, hypervisor, cache.NoExpiration)
	return hypervisor, nil
}

// fetchNodeData is a helper method that fetches nodes data over rmb
// returns the node capacity, hypervisor and dmi
func (a *App) fetchNodeData(nodeID string) (NodeInfo, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*30)
	defer cancel()

	twinID, err := a.getNodeTwinID(nodeID)
	if err != nil {
		return NodeInfo{}, err
	}

	capacity, err := a.getNodeCapacity(ctx, nodeID, false)
	if err != nil {
		return NodeInfo{}, errors.Wrapf(err, "error fetching node capacity")
	}

	nodeClient := client.NewNodeClient(twinID, a.rmb)
	// get node version
	version, err := nodeClient.SystemVersion(ctx)
	if err != nil {
		return NodeInfo{}, errors.Wrapf(err, "error fetching node version")
	}

	// get node hypervisor
	hypervisor, err := a.getNodeHypervisor(ctx, nodeID, nodeClient)
	if err != nil {
		return NodeInfo{}, errors.Wrapf(err, "error fetching hypervisor")
	}

	// get node dmi

	log.Debug().Str("duration", time.Since(start).String()).Msg("time taken to fetch node data")
	return NodeInfo{
		Capacity:   capacity,
		Hypervisor: hypervisor,
		ZosVersion: version.ZOS,
	}, nil
}

// getNodeData is a helper function that wraps fetch node data
// it caches the results in redis to save time
func (a *App) getNodeData(nodeID string, force bool) (string, error) {
	value, _ := a.GetRedisKey(a.getNodeKey(nodeID))

	// value exists just return it
	if value != "" && !force {
		return value, nil
	}

	nodeInfo, fetchingNodesError := a.fetchNodeData(nodeID)
	if errors.Is(fetchingNodesError, ErrNodeNotFound) {
		// delete redis key
		err := a.DeleteRedisKey(a.getNodeKey(nodeID))
		if err != nil {
			log.Warn().Err(err).Msg("could not delete key in redis")
		}
		return "", ErrNodeNotFound
	} else if fetchingNodesError != nil {
		// if node is down delete the key and return bad gateway
		err := a.DeleteRedisKey(a.getNodeKey(nodeID))
		if err != nil {
			log.Warn().Err(err).Msg("could not delete key in redis")
		}
		return "", errors.Wrapf(ErrBadGateway, fetchingNodesError.Error())
	}
	// Save value in redis
	// caching for 30 mins
	serializedNodeInfo, err := nodeInfo.Serialize()
	if err != nil {
		return "", err
	}

	err = a.SetRedisKey(a.getNodeKey(nodeID), serializedNodeInfo, 30*60)
	if err != nil {
		log.Warn().Err(err).Msg("could not cache data in redis")
	}
	return string(serializedNodeInfo), nil
}

// getNodeCapacity is a helper function that wraps fetch node capacity
// it caches the results in redis to save time
func (a *App) getNodeCapacity(ctx context.Context, nodeID string, force bool) (capacityResult, error) {
	value, _ := a.GetRedisKey(a.getNodeCapacityKey(nodeID))
	nodeCapacityResult := capacityResult{}

	// value exists just return it
	if value != "" && !force {
		_ = json.Unmarshal([]byte(value), &nodeCapacityResult)
		return nodeCapacityResult, nil
	}

	twinID, err := a.getNodeTwinID(nodeID)
	if err != nil {
		return capacityResult{}, err
	}

	nodeClient := client.NewNodeClient(twinID, a.rmb)
	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	// get node capacity
	total, used, err := nodeClient.Counters(ctx2)
	if err != nil {
		return nodeCapacityResult, errors.Wrapf(err, "error fetching node statistics")
	}
	nodeCapacityResult.Total = total
	nodeCapacityResult.Used = used

	// Save value in redis
	// caching for 30 mins
	serializedNodeInfo, err := json.Marshal(nodeCapacityResult)
	if err != nil {
		log.Warn().Err(err).Msg("could not cache data in redis")
		return nodeCapacityResult, nil
	}

	err = a.SetRedisKey(a.getNodeCapacityKey(nodeID), serializedNodeInfo, 10*60)
	if err != nil {
		log.Warn().Err(err).Msg("could not cache data in redis")
	}
	return nodeCapacityResult, nil
}

// getAllNodesIDs is a helper method to only list all nodes ids
func (a *App) getAllNodesIDs() (nodeIDResult, error) {
	queryString := `
	{
		nodes(limit:99999999){
			nodeId
		}    
	}
	`
	nodesIds := nodeIDResult{}
	err := a.query(queryString, &nodesIds)
	if err != nil {
		return nodeIDResult{}, fmt.Errorf("failed to query nodes %w", err)
	}
	return nodesIds, nil
}

// cacheNodesInfo is a helper method that caches nodes data into redis
// it runs at the begining of the application and every 2 mins
func (a *App) cacheNodesInfo() {
	nodeIds, err := a.getAllNodesIDs()
	if err != nil {
		log.Error().Err(err).Msg("failed to query nodes")
		return
	}

	channelLimit := make(chan int, maxGoRoutnes)
	defer close(channelLimit)
	for i, nid := range nodeIds.Data.NodeResult {
		channelLimit <- 1
		go func(i int, nid nodeID) {
			log.Debug().Msg(fmt.Sprintf("%d:fetching node: %d", i+1, nid.NodeID))
			_, err := a.getNodeData(fmt.Sprint(nid.NodeID), true)
			if err != nil {
				log.Warn().Err(err).Msg(fmt.Sprintf("could not fetch node data %d", nid.NodeID))
			} else {
				log.Debug().Msg(fmt.Sprintf("node %d is fetched successfully", nid.NodeID))
			}
			<-channelLimit
		}(i, nid)
	}
	log.Debug().Msg("Fetching nodes completed, next fetch will be in 15 minutes")
}

func (a *App) getTotalCount() (int, error) {
	queryString := `
	{
		nodesConnection {
			totalCount
		}
	}	
	`
	nodesCount := totalCountResponse{}
	err := a.query(queryString, &nodesCount)
	if err != nil {
		return 0, fmt.Errorf("failed to query nodes %w", err)
	}
	return nodesCount.Date.NodesConnection.TotalCount, nil
}


// getAllNodes is a helper method to list all nodes data and set it to the proper struct
func (a *App) getAllNodes(maxResult int, pageOffset int, isSpecificFarm string, isGateway string) (nodesResponse, error) {

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
			certificationType
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

	nodes := nodesResponse{}
	err := a.query(queryString, &nodes)
	if err != nil {
		return nodesResponse{}, fmt.Errorf("failed to query nodes %w", err)
	}
	return nodes, nil
}
