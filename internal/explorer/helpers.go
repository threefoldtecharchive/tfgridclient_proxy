package explorer

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/threefoldtech/grid_proxy_server/internal/explorer/db"
	"github.com/threefoldtech/grid_proxy_server/internal/explorer/mw"
)

const (
	maxPageSize = 100
)

func errorReply(err error) mw.Response {
	if errors.Is(err, ErrNodeNotFound) {
		return mw.NotFound(err)
	} else if errors.Is(err, ErrBadGateway) {
		return mw.BadGateway(err)
	} else {
		return mw.Error(err)
	}
}

func getLimit(r *http.Request) (db.Limit, error) {
	var limit db.Limit

	page := r.URL.Query().Get("page")
	size := r.URL.Query().Get("size")
	retCount := r.URL.Query().Get("ret_count")
	if page == "" {
		page = "1"
	}
	if size == "" {
		size = "50"
	}
	parsed, err := strconv.ParseUint(page, 10, 64)
	if err != nil {
		return limit, errors.Wrap(ErrBadRequest, fmt.Sprintf("couldn't parse page %s", err.Error()))
	}
	limit.Page = parsed

	parsed, err = strconv.ParseUint(size, 10, 64)
	if err != nil {
		return limit, errors.Wrap(ErrBadRequest, fmt.Sprintf("couldn't parse size %s", err.Error()))
	}
	limit.Size = parsed

	returnCount := false
	if retCount == "true" {
		returnCount = true
	}
	limit.RetCount = returnCount

	// TODO: readd the check once clients are updated
	// if limit.Size > maxPageSize {
	// 	return limit, errors.Wrapf(ErrBadRequest, "max page size is %d", maxPageSize)
	// }
	return limit, nil
}
func parseParams(
	r *http.Request,
	ints map[string]**uint64,
	strs map[string]**string,
	bools map[string]**bool,
	listOfInts map[string]*[]uint64,
) error {
	for param, prop := range ints {
		value := r.URL.Query().Get(param)
		if value != "" {
			parsed, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return errors.Wrap(ErrBadRequest, fmt.Sprintf("couldn't parse %s %s", param, err.Error()))
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
	falseVal := false
	for param, prop := range bools {
		value := r.URL.Query().Get(param)
		if value == "true" {
			*prop = &trueVal
		}
		if value == "false" {
			*prop = &falseVal
		}
	}
	for param, prop := range listOfInts {
		value := r.URL.Query().Get(param)
		if value == "" {
			continue
		} else {
			split := strings.Split(value, ",")
			*prop = make([]uint64, 0)
			for _, item := range split {
				parsed, err := strconv.ParseUint(item, 10, 64)
				if err != nil {
					return errors.Wrap(ErrBadRequest, fmt.Sprintf("couldn't parse %s %s", param, err.Error()))
				}
				*prop = append(*prop, parsed)
			}
		}
	}
	return nil
}

// test nodes?status=up&free_ips=0&free_cru=1&free_mru=1&free_hru=1&country=Belgium&city=Unknown&ipv4=true&ipv6=true&domain=false
// HandleNodeRequestsQueryParams takes the request and restore the query paramas, handle errors and set default values if not available
func (a *App) handleNodeRequestsQueryParams(r *http.Request) (db.NodeFilter, db.Limit, error) {
	var filter db.NodeFilter
	var limit db.Limit
	ints := map[string]**uint64{
		"free_mru":      &filter.FreeMRU,
		"free_hru":      &filter.FreeHRU,
		"free_sru":      &filter.FreeSRU,
		"free_ips":      &filter.FreeIPs,
		"rented_by":     &filter.RentedBy,
		"available_for": &filter.AvailableFor,
	}
	strs := map[string]**string{
		"status":    &filter.Status,
		"city":      &filter.City,
		"country":   &filter.Country,
		"farm_name": &filter.FarmName,
	}
	bools := map[string]**bool{
		"ipv4":     &filter.IPv4,
		"ipv6":     &filter.IPv6,
		"domain":   &filter.Domain,
		"rentable": &filter.Rentable,
	}
	listOfInts := map[string]*[]uint64{
		"farm_ids": &filter.FarmIDs,
	}
	if err := parseParams(r, ints, strs, bools, listOfInts); err != nil {
		return filter, limit, err
	}
	limit, err := getLimit(r)
	if err != nil {
		return filter, limit, err
	}
	trueval := true
	if strings.HasSuffix(r.URL.Path, "gateways") {
		filter.Domain = &trueval
		filter.IPv4 = &trueval
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
		"total_ips":         &filter.TotalIPs,
		"pricing_policy_id": &filter.PricingPolicyID,
		"version":           &filter.Version,
		"farm_id":           &filter.FarmID,
		"twin_id":           &filter.TwinID,
	}
	strs := map[string]**string{
		"name":               &filter.Name,
		"name_contains":      &filter.NameContains,
		"certification_type": &filter.CertificationType,
		"stellar_address":    &filter.StellarAddress,
	}
	bools := map[string]**bool{
		"dedicated": &filter.Dedicated,
	}
	if err := parseParams(r, ints, strs, bools, nil); err != nil {
		return filter, limit, err
	}

	limit, err := getLimit(r)
	if err != nil {
		return filter, limit, err
	}
	return filter, limit, nil
}

// test stats?status=up
// HandleNodeRequestsQueryParams takes the request and restore the query paramas, handle errors and set default values if not available
func (a *App) handleStatsRequestsQueryParams(r *http.Request) (db.StatsFilter, error) {
	var filter db.StatsFilter
	strs := map[string]**string{
		"status": &filter.Status,
	}
	if err := parseParams(r, nil, strs, nil, nil); err != nil {
		return filter, err
	}
	return filter, nil
}

func (a *App) getTotalCount() (uint, error) {
	return a.db.CountNodes()
}

// getNodeData is a helper function that wraps fetch node data
// it caches the results in redis to save time
func (a *App) getNodeData(nodeIDStr string) (nodeWithNestedCapacity, error) {
	nodeID, err := strconv.Atoi(nodeIDStr)
	if err != nil {
		return nodeWithNestedCapacity{}, errors.Wrap(ErrBadGateway, fmt.Sprintf("invalid node id %d: %s", nodeID, err.Error()))
	}
	info, err := a.db.GetNode(uint32(nodeID))
	if errors.Is(err, db.ErrNodeNotFound) {
		return nodeWithNestedCapacity{}, ErrNodeNotFound
	} else if err != nil {
		// TODO: wrapping
		return nodeWithNestedCapacity{}, err
	}
	apiNode := nodeWithNestedCapacityFromDBNode(info)
	return apiNode, nil
}
