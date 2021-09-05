package explorer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

func getNodeTwinID(nodeID string, explorerURL string) (uint32, error) {
	queryString := fmt.Sprintf(`
	{
		nodes(limit:10, where:{nodeId_eq:%s}){
		  twinId
		}
	}
	`, nodeID)

	var res nodeResult
	err := query(queryString, explorerURL, &res)

	if err != nil {
		return 0, fmt.Errorf("failed to query node %w", err)
	}

	nodeStats := res.Data.NodeResult
	if len(nodeStats) > 0 {
		return nodeStats[0].TwinID, nil
	}
	return 0, fmt.Errorf("failed to find node ID")
}

func baseQuery(queryString string, explorerURL string) (io.ReadCloser, error) {
	jsonData := map[string]string{
		"query": queryString,
	}
	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		return nil, fmt.Errorf("invalid query string %w", err)
	}

	request, err := http.NewRequest("POST", explorerURL, bytes.NewBuffer(jsonValue))
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

func query(queryString string, explorerURL string, result interface{}) error {
	response, err := baseQuery(queryString, explorerURL)
	if err != nil {
		return err
	}

	defer response.Close()
	if err := json.NewDecoder(response).Decode(result); err != nil {
		return err
	}

	return nil
}

func queryProxy(queryString string, explorerURL string, w io.Writer) (written int64, err error) {
	response, err := baseQuery(queryString, explorerURL)
	if err != nil {
		return 0, err
	}

	defer response.Close()

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
		page = "0"
	}

	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		return 0, fmt.Errorf("invalid page number : %w", err)
	}

	offset := 0
	if pageNumber > 1 {
		offset = pageNumber * maxResult
	}

	return offset, nil
}

// HandleRequestsQueryParams takes the request and restore the query paramas, handle errors and set default values if not available
func (a *App) handleRequestsQueryParams(r *http.Request) (*http.Request, error) {

	farmID := r.URL.Query().Get("farm_id")
	isSpecificFarm := ""
	if farmID != "" {
		isSpecificFarm = fmt.Sprintf(",where:{farmId_eq:%s}", farmID)
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

	return r.WithContext(ctx), nil
}
