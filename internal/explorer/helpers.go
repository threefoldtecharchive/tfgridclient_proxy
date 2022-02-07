package explorer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/grid_proxy_server/internal/explorer/db"
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

// HandleNodeRequestsQueryParams takes the request and restore the query paramas, handle errors and set default values if not available
func (a *App) handleNodeRequestsQueryParams(r *http.Request) (db.NodeFilter, db.Limit, error) {
	var filter db.NodeFilter
	var limit db.Limit

	freeCRU := r.URL.Query().Get("free_cru")
	if freeCRU != "" {
		parsed, err := strconv.ParseUint(freeCRU, 10, 64)
		if err != nil {
			return filter, limit, errors.Wrap(ErrBadGateway, fmt.Sprintf("couldn't parse free_cru %s", err.Error()))
		}
		filter.FreeCRU = &parsed
	}
	freeMRU := r.URL.Query().Get("free_mru")
	if freeMRU != "" {
		parsed, err := strconv.ParseUint(freeMRU, 10, 64)
		if err != nil {
			return filter, limit, errors.Wrap(ErrBadGateway, fmt.Sprintf("couldn't parse free_cu %s", err.Error()))
		}
		filter.FreeMRU = &parsed
	}
	freeHRU := r.URL.Query().Get("free_hru")
	if freeHRU != "" {
		parsed, err := strconv.ParseUint(freeHRU, 10, 64)
		if err != nil {
			return filter, limit, errors.Wrap(ErrBadGateway, fmt.Sprintf("couldn't parse free_hru %s", err.Error()))
		}
		filter.FreeHRU = &parsed
	}
	freeSRU := r.URL.Query().Get("free_sru")
	if freeSRU != "" {
		parsed, err := strconv.ParseUint(freeSRU, 10, 64)
		if err != nil {
			return filter, limit, errors.Wrap(ErrBadGateway, fmt.Sprintf("couldn't parse free_sru %s", err.Error()))
		}
		filter.FreeSRU = &parsed
	}

	status := r.URL.Query().Get("status")
	if status != "" {
		filter.Status = &status
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
		return filter, limit, errors.Wrap(ErrBadGateway, fmt.Sprintf("couldn't parse page %s", err.Error()))
	}
	limit.Page = parsed

	parsed, err = strconv.ParseUint(size, 10, 64)
	if err != nil {
		return filter, limit, errors.Wrap(ErrBadGateway, fmt.Sprintf("couldn't parse size %s", err.Error()))
	}
	limit.Size = parsed
	return filter, limit, nil
}

// HandleFarmRequestsQueryParams takes the request and restore the query paramas, handle errors and set default values if not available
func (a *App) handleFarmRequestsQueryParams(r *http.Request) (db.FarmFilter, db.Limit, error) {
	var filter db.FarmFilter
	var limit db.Limit

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
		return filter, limit, errors.Wrap(ErrBadGateway, fmt.Sprintf("couldn't parse page %s", err.Error()))
	}
	limit.Page = parsed

	parsed, err = strconv.ParseUint(size, 10, 64)
	if err != nil {
		return filter, limit, errors.Wrap(ErrBadGateway, fmt.Sprintf("couldn't parse size %s", err.Error()))
	}
	limit.Size = parsed
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
