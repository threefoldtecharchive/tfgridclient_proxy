package explorer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/grid_proxy_server/internal/explorer/db"
)

const (
	maxPageSize = 100
)

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

// test nodes?status=up&free_ips=0&free_cru=1&free_mru=1&free_hru=1&country=Belgium&city=Unknown&ipv4=true&ipv6=true&domain=false
// HandleNodeRequestsQueryParams takes the request and restore the query paramas, handle errors and set default values if not available
func (a *App) handleNodeRequestsQueryParams(r *http.Request) (db.NodeFilter, db.Limit, error) {
	var filter db.NodeFilter
	var limit db.Limit
	ints := map[string]**uint64{
		"free_cru": &filter.FreeCRU,
		"free_mru": &filter.FreeMRU,
		"free_hru": &filter.FreeHRU,
		"free_sru": &filter.FreeSRU,
		"free_ips": &filter.FreeIPs,
	}
	strs := map[string]**string{
		"status":    &filter.Status,
		"city":      &filter.City,
		"country":   &filter.Country,
		"farm_name": &filter.FarmName,
	}
	bools := map[string]**bool{
		"ipv4":   &filter.IPv4,
		"ipv6":   &filter.IPv6,
		"domain": &filter.Domain,
	}
	for param, prop := range ints {
		value := r.URL.Query().Get(param)
		if value != "" {
			parsed, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return filter, limit, errors.Wrap(ErrBadRequest, fmt.Sprintf("couldn't parse %s %s", param, err.Error()))
			}
			*prop = &parsed
		}
	}

	for param, prop := range strs {
		value := r.URL.Query().Get(param)
		if value != "" {
			*prop = &value
		}
	}
	trueVal := true
	for param, prop := range bools {
		value := r.URL.Query().Get(param)
		if value == "true" {
			*prop = &trueVal
		}
	}
	farmIDs := strings.Split(r.URL.Query().Get("farm_ids"), ",")
	if len(farmIDs) != 1 || farmIDs[0] != "" {
		filter.FarmIDs = make([]uint64, len(farmIDs))
		for idx, id := range farmIDs {
			parsed, err := strconv.ParseUint(strings.TrimSpace(id), 10, 64)
			if err != nil {
				return filter, limit, errors.Wrap(ErrBadRequest, fmt.Sprintf("couldn't parse farm_id indexed at %d %s", idx, err.Error()))
			}
			filter.FarmIDs[idx] = parsed
		}
	}
	page := r.URL.Query().Get("page")
	size := r.URL.Query().Get("size")
	if page == "" {
		page = "1"
	}
	if size == "" {
		size = "50"
	}
	parsed, err := strconv.ParseUint(page, 10, 64)
	if err != nil {
		return filter, limit, errors.Wrap(ErrBadRequest, fmt.Sprintf("couldn't parse page %s", err.Error()))
	}
	limit.Page = parsed

	parsed, err = strconv.ParseUint(size, 10, 64)
	if err != nil {
		return filter, limit, errors.Wrap(ErrBadRequest, fmt.Sprintf("couldn't parse size %s", err.Error()))
	}
	limit.Size = parsed
	if limit.Size >= maxPageSize {
		return filter, limit, errors.Wrapf(ErrBadRequest, "max page size is %d", maxPageSize)
	}
	return filter, limit, nil
}

// test farms?free_ips=1&pricing_policy_id=1&version=4&farm_id=23&twin_id=291&name=Farm-1&stellar_address=13VrxhaBZh87ZP8nuYF4LtAhnDPWMfSrMUvHeRAFaqN43W1X
// HandleFarmRequestsQueryParams takes the request and restore the query paramas, handle errors and set default values if not available
func (a *App) handleFarmRequestsQueryParams(r *http.Request) (db.FarmFilter, db.Limit, error) {
	var filter db.FarmFilter
	var limit db.Limit

	ints := map[string]**uint64{
		"free_ips":          &filter.FreeIPs,
		"pricing_policy_id": &filter.PricingPolicyID,
		"version":           &filter.Version,
		"farm_id":           &filter.FarmID,
		"twin_id":           &filter.TwinID,
	}
	strs := map[string]**string{
		"name":            &filter.Name,
		"stellar_address": &filter.StellarAddress,
	}
	for param, prop := range ints {
		value := r.URL.Query().Get(param)
		if value != "" {
			parsed, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return filter, limit, errors.Wrap(ErrBadRequest, fmt.Sprintf("couldn't parse %s %s", param, err.Error()))
			}
			*prop = &parsed
		}
	}

	for param, prop := range strs {
		value := r.URL.Query().Get(param)
		if value != "" {
			*prop = &value
		}
	}

	page := r.URL.Query().Get("page")
	size := r.URL.Query().Get("size")
	if page == "" {
		page = "1"
	}
	if size == "" {
		size = "50"
	}
	parsed, err := strconv.ParseUint(page, 10, 64)
	if err != nil {
		return filter, limit, errors.Wrap(ErrBadRequest, fmt.Sprintf("couldn't parse page %s", err.Error()))
	}
	limit.Page = parsed

	parsed, err = strconv.ParseUint(size, 10, 64)
	if err != nil {
		return filter, limit, errors.Wrap(ErrBadRequest, fmt.Sprintf("couldn't parse size %s", err.Error()))
	}
	limit.Size = parsed
	if limit.Size >= maxPageSize {
		return filter, limit, errors.Wrapf(ErrBadRequest, "max page size is %d", maxPageSize)
	}
	return filter, limit, nil
}

// getNodeData is a helper function that wraps fetch node data
// it caches the results in redis to save time
func (a *App) getNodeData(nodeIDStr string) (node, error) {
	nodeID, err := strconv.Atoi(nodeIDStr)
	if err != nil {
		return node{}, errors.Wrap(ErrBadGateway, fmt.Sprintf("invalid node id %d: %s", nodeID, err.Error()))
	}
	info, err := a.db.GetNode(uint32(nodeID))
	if errors.Is(err, db.ErrNodeNotFound) {
		return node{}, ErrNodeNotFound
	} else if err != nil {
		// TODO: wrapping
		return node{}, err
	}
	apiNode := nodeFromDBNode(info)
	return apiNode, nil
}
