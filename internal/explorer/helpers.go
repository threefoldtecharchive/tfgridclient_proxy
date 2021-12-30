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
	if len(nodeStats) > 0 {
		twinID := nodeStats[0].TwinID
		a.lruCache.Set(nodeID, twinID, cache.DefaultExpiration)
		return twinID, nil
	}
	return 0, ErrNodeNotFound
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
	if response.StatusCode != 200 {
		var errResult interface{}
		if err := json.NewDecoder(response.Body).Decode(errResult); err != nil {
			return nil, fmt.Errorf("request failed: %w", errResult)
		}
		return nil, fmt.Errorf("failed to query explorer network")
	}
	return response.Body, nil
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

func (a *App) queryProxy(queryString string, w http.ResponseWriter) (written int64, err error) {
	response, err := a.baseQuery(queryString)
	if err != nil {
		return 0, err
	}

	defer response.Close()
	w.Header().Add("Content-Type", "application/json")
	return io.Copy(w, response)
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
	maxResultPerpage := r.URL.Query().Get("max_result")
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

func (a *App) getNodeCapacity(ctx context.Context, nodeClient *client.NodeClient, c chan countersReturnValue) {
	cap, used, err := nodeClient.Counters(ctx)
	capacity := countersReturnValue{
		NodeCapacity: cap,
		UsedCapacity: used,
		Err:          err,
	}
	c <- capacity
}

func (a *App) getSystemDMI(ctx context.Context, nodeClient *client.NodeClient, c chan systemDMIReturnValue) {
	dmi, err := nodeClient.SystemDMI(ctx)
	systemDMI := systemDMIReturnValue{
		DMI: dmi,
		Err: err,
	}
	c <- systemDMI
}

func (a *App) getSystemHypervisor(ctx context.Context, nodeClient *client.NodeClient, c chan systemHypervisorReturnValue) {
	hypervisor, err := nodeClient.SystemHypervisor(ctx)
	systemHypervisor := systemHypervisorReturnValue{
		Hypervisor: hypervisor,
		Err:        err,
	}
	c <- systemHypervisor
}

func (a *App) getSystemVersion(ctx context.Context, nodeClient *client.NodeClient, c chan systemVersionReturnValue) {
	version, err := nodeClient.SystemVersion(ctx)
	systemVersion := systemVersionReturnValue{
		ZosVersion: version,
		Err:        err,
	}
	c <- systemVersion
}

// fetchNodeData is a helper method that fetches nodes data over rmb
// returns the node capacity, hypervisor and dmi
func (a *App) fetchNodeData(nodeID string) (NodeInfo, error) {
	twinID, err := a.getNodeTwinID(nodeID)
	if err != nil {
		return NodeInfo{}, err

	}
	ctx := context.Background()

	nodeClient := client.NewNodeClient(twinID, a.rmb)
	nodeCapacity := make(chan countersReturnValue)

	go a.getNodeCapacity(ctx, nodeClient, nodeCapacity)

	systemDMI := make(chan systemDMIReturnValue)
	go a.getSystemDMI(ctx, nodeClient, systemDMI)

	systemHypervisor := make(chan systemHypervisorReturnValue)
	go a.getSystemHypervisor(ctx, nodeClient, systemHypervisor)

	systemVersion := make(chan systemVersionReturnValue)
	go a.getSystemVersion(ctx, nodeClient, systemVersion)

	p := <-nodeCapacity
	d := <-systemDMI
	h := <-systemHypervisor
	v := <-systemVersion

	if p.Err != nil {
		return NodeInfo{}, fmt.Errorf("error fetching node data : %w", p.Err)
	}
	if d.Err != nil {
		return NodeInfo{}, fmt.Errorf("error fetching node data : %w", d.Err)
	}
	if h.Err != nil {
		return NodeInfo{}, fmt.Errorf("error fetching node data : %w", h.Err)
	}
	if v.Err != nil {
		return NodeInfo{}, fmt.Errorf("error fetching node data : %w", v.Err)
	}

	capacity := capacityResult{}
	capacity.Total = p.NodeCapacity
	capacity.Used = p.UsedCapacity

	return NodeInfo{
		Capacity:   capacity,
		DMI:        d.DMI,
		Hypervisor: h.Hypervisor,
		ZosVersion: v.ZosVersion.ZOS,
	}, nil

}

// getNodeData is a helper function that wraps fetch node data
// it caches the results in redis to save time
func (a *App) getNodeData(nodeID string, force bool) (string, error) {
	value, _ := a.GetRedisKey(fmt.Sprintf("GRID3NODE:%s", nodeID))

	// No value, fetch data from the node
	if value == "" || force {
		nodeInfo, err := a.fetchNodeData(nodeID)
		if errors.Is(err, ErrNodeNotFound) {
			// delete redis key
			err = a.DeleteRedisKey(fmt.Sprintf("GRID3NODE:%s", nodeID))
			if err != nil {
				log.Warn().Err(err).Msg("could not delete key in redis")
			}
			return "", ErrNodeNotFound
		} else if err != nil {
			err = a.DeleteRedisKey(fmt.Sprintf("GRID3NODE:%s", nodeID))
			if err != nil {
				log.Warn().Err(err).Msg("could not delete key in redis")
			}
			return "", ErrBadGateway
		}
		// Save value in redis
		// caching for 40 mins
		marshalledInfo, err := json.Marshal(nodeInfo)
		if err != nil {
			log.Error().Err(err).Msg("could not marshal node info")
			return "", errors.Wrap(err, "internal server error")
		}

		err = a.SetRedisKey(fmt.Sprintf("GRID3NODE:%s", nodeID), marshalledInfo, 40*60)
		if err != nil {
			log.Warn().Err(err).Msg("could not cache data in redis")
		}
		value = string(marshalledInfo)
	}
	return value, nil
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
	}

	for i, nid := range nodeIds.Data.NodeResult {
		go func(i int, nid nodeID) {
			log.Debug().Msg(fmt.Sprintf("%d:fetching node: %d", i+1, nid.NodeID))
			_, err := a.getNodeData(fmt.Sprint(nid.NodeID), true)
			if err != nil {
				log.Error().Err(err).Msg(fmt.Sprintf("could not fetch node data %d", nid.NodeID))
			} else {
				log.Debug().Msg(fmt.Sprintf("node %d is fetched successfully", nid.NodeID))
			}
		}(i, nid)
	}
	log.Debug().Msg("Fetching nodes completed, next fetch will be in 15 minutes")
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
