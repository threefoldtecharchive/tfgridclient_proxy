package explorer

import (
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"

	// swagger configuration
	_ "github.com/threefoldtech/grid_proxy_server/docs"
	"github.com/threefoldtech/grid_proxy_server/internal/explorer/db"
	"github.com/threefoldtech/grid_proxy_server/internal/explorer/mw"
	"github.com/threefoldtech/grid_proxy_server/pkg/types"
	"github.com/threefoldtech/zos/pkg/rmb"
)

const (
	// SSDOverProvisionFactor factor by which the ssd are allowed to be overprovisioned
	SSDOverProvisionFactor = 2
)

func (a *App) listFarms(r *http.Request) (interface{}, mw.Response) {
	filter, limit, err := a.handleFarmRequestsQueryParams(r)
	if err != nil {
		return nil, mw.BadRequest(err)
	}
	dbFarms, farmsCount, err := a.db.GetFarms(filter, limit)
	if err != nil {
		log.Error().Err(err).Msg("failed to query farm")
		return nil, mw.Error(err)
	}
	farms := make([]types.Farm, 0, len(dbFarms))
	for _, farm := range dbFarms {
		f, err := farmFromDBFarm(farm)
		if err != nil {
			log.Err(err).Msg("couldn't convert db farm to api farm")
		}
		farms = append(farms, f)
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

func (a *App) loadStats(r *http.Request) (interface{}, mw.Response) {
	filter, err := a.handleStatsRequestsQueryParams(r)
	if err != nil {
		return nil, mw.BadRequest(err)
	}
	counters, err := a.db.GetCounters(filter)
	if err != nil {
		return nil, mw.Error(err)
	}
	return counters, nil
}

func (a *App) listNodes(r *http.Request) (interface{}, mw.Response) {
	filter, limit, err := a.handleNodeRequestsQueryParams(r)
	if err != nil {
		return nil, mw.BadRequest(err)
	}
	dbNodes, nodesCount, err := a.db.GetNodes(filter, limit)
	if err != nil {
		return nil, mw.Error(err)
	}
	nodes := make([]types.Node, len(dbNodes))
	for idx, node := range dbNodes {
		nodes[idx] = nodeFromDBNode(node)
	}
	resp := mw.Ok()

	// return the number of pages and totalCount in the response headers
	if limit.RetCount {
		pages := math.Ceil(float64(nodesCount) / float64(limit.Size))
		resp = resp.WithHeader("count", fmt.Sprintf("%d", nodesCount)).
			WithHeader("size", fmt.Sprintf("%d", limit.Size)).
			WithHeader("pages", fmt.Sprintf("%d", int(pages)))
	}
	return nodes, resp
}

func (a *App) loadNode(r *http.Request) (interface{}, mw.Response) {
	nodeID := mux.Vars(r)["node_id"]
	nodeData, err := a.getNodeData(nodeID)
	if err != nil {
		return nil, errorReply(err)
	}
	return nodeData, nil
}

func (a *App) listTwins(r *http.Request) (interface{}, mw.Response) {
	filter, limit, err := a.handleTwinRequestsQueryParams(r)
	if err != nil {
		return nil, mw.BadRequest(err)
	}
	twins, twinsCount, err := a.db.GetTwins(filter, limit)
	if err != nil {
		log.Error().Err(err).Msg("failed to query twin")
		return nil, mw.Error(err)
	}

	resp := mw.Ok()

	// return the number of pages and totalCount in the response headers
	if limit.RetCount {
		pages := math.Ceil(float64(twinsCount) / float64(limit.Size))
		resp = resp.WithHeader("count", fmt.Sprintf("%d", twinsCount)).
			WithHeader("size", fmt.Sprintf("%d", limit.Size)).
			WithHeader("pages", fmt.Sprintf("%d", int(pages)))
	}
	return twins, resp
}

func (a *App) listContracts(r *http.Request) (interface{}, mw.Response) {
	filter, limit, err := a.handleContractRequestsQueryParams(r)
	if err != nil {
		return nil, mw.BadRequest(err)
	}
	dbContracts, contractsCount, err := a.db.GetContracts(filter, limit)
	if err != nil {
		log.Error().Err(err).Msg("failed to query contract")
		return nil, mw.Error(err)
	}

	contracts := make([]types.Contract, len(dbContracts))
	for idx, contract := range dbContracts {
		contracts[idx], err = contractFromDBContract(contract)
		if err != nil {
			log.Err(err).Msg("failed to convert db contract to api contract")
		}
	}
	resp := mw.Ok()

	// return the number of pages and totalCount in the response headers
	if limit.RetCount {
		pages := math.Ceil(float64(contractsCount) / float64(limit.Size))
		resp = resp.WithHeader("count", fmt.Sprintf("%d", contractsCount)).
			WithHeader("size", fmt.Sprintf("%d", limit.Size)).
			WithHeader("pages", fmt.Sprintf("%d", int(pages)))
	}
	return contracts, resp
}

func (a *App) indexPage(m *mux.Router) mw.Action {
	return func(r *http.Request) (interface{}, mw.Response) {
		response := mw.Ok()
		var sb strings.Builder
		sb.WriteString("Welcome to threefold grid proxy server, available endpoints ")

		_ = m.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			path, err := route.GetPathTemplate()
			if err != nil {
				return nil
			}

			sb.WriteString("[" + path + "] ")
			return nil
		})
		return sb.String(), response
	}
}

func (a *App) Version(r *http.Request) (interface{}, mw.Response) {
	response := mw.Ok()
	return types.Version{
		Version: a.releaseVersion,
	}, response
}

func (a *App) GetNodeStatus(r *http.Request) (interface{}, mw.Response) {
	response := types.NodeStatus{}
	nodeID := mux.Vars(r)["node_id"]

	nodeData, err := a.getNodeData(nodeID)
	if err != nil {
		return nil, errorReply(err)
	}
	response.Status = nodeData.Status
	return response, nil
}

func Setup(version string, router *mux.Router, rmbClient rmb.Client, c *cache.Cache, gitCommit string, database db.Database) error {
	a := App{
		db:             database,
		rmb:            rmbClient,
		lruCache:       c,
		releaseVersion: gitCommit,
	}

	if version == "v1" {
		a.loadV1Handlers(router)
	} else {
		a.loadV2Handlers(router)
	}

	return nil
}
