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

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/zos/pkg/gridtypes"
	"github.com/threefoldtech/zos/pkg/rmb"
)

// URL is the default explorer graphql url
const URL string = "https://explorer.devnet.grid.tf/graphql/"

// NewNodeClient Creates new node client from the twin id
func NewNodeClient(nodeTwin uint32, bus rmb.Client) *NodeClient {
	return &NodeClient{nodeTwin, bus}
}

func getNodeTwinID(nodeID string) (uint32, error) {
	queryString := fmt.Sprintf(`
	{
		nodes(limit:10, where:{nodeId_eq:%s}){
		  twinId
		}
	}
	`, nodeID)

	var res NodeResult
	err := query(queryString, &res)

	if err != nil {
		log.Error().Err(errors.Wrap(err, "couldn't parse json")).Msg("connection error")
		return 0, fmt.Errorf("error: couldn't get node twinID %w", err)
	}

	nodeStats := res.Data.NodeResult
	if len(nodeStats) > 0 {
		log.Info().Str("Node twin id", fmt.Sprint(nodeStats[0].TwinID)).Msg("Preparing Node data")
		return nodeStats[0].TwinID, nil
	}
	return 0, fmt.Errorf("failed to find node ID")

}

// NodeStatistics Returns actual node Statistics from the node itself over the msgbus
func (n *NodeClient) NodeStatistics(ctx context.Context) (total CapacityResult, err error) {
	const cmd = "zos.statistics.get"
	var result struct {
		Total gridtypes.Capacity `json:"total"`
		Used  gridtypes.Capacity `json:"used"`
	}
	if err = n.bus.Call(ctx, n.nodeTwin, cmd, nil, &result); err != nil {
		return
	}

	return result, nil
}

func baseQuery(queryString string) (io.ReadCloser, error) {
	jsonData := map[string]string{
		"query": queryString,
	}
	jsonValue, _ := json.Marshal(jsonData)
	request, err := http.NewRequest("POST", URL, bytes.NewBuffer(jsonValue))
	request.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Error().Err(errors.Wrap(err, "Failed to connect to graphql network")).Msg("connection error")
	}

	client := &http.Client{Timeout: time.Second * 10}
	response, err := client.Do(request)
	if err != nil {
		log.Error().Err(errors.Wrap(err, "Failed to connect to graphql network")).Msg("connection error")
	}
	return response.Body, err
}

func query(queryString string, result interface{}) error {
	response, err := baseQuery(queryString)
	if err != nil {
		log.Error().Err(errors.Wrap(err, "Failed to connect to graphql network")).Msg("connection error")
	}
	defer response.Close()
	if err := json.NewDecoder(response).Decode(result); err != nil {
		return err
	}
	return nil
}

func queryProxy(queryString string, w io.Writer) (written int64, err error) {
	response, err := baseQuery(queryString)
	if err != nil {
		log.Error().Err(errors.Wrap(err, "Failed to connect to graphql network")).Msg("connection error")
	}
	defer response.Close()
	return io.Copy(w, response)
}

// GetOffset is helper function to get offest from context
func GetOffset(ctx context.Context) int {
	return ctx.Value(OffsetKey{}).(int)
}

// GetMaxResult is helper function to get MaxResult from context
func GetMaxResult(ctx context.Context) int {
	return ctx.Value(MaxResultKey{}).(int)
}

// GetSpecificFarm is helper function to get SpecificFarm from context
func GetSpecificFarm(ctx context.Context) string {
	return ctx.Value(SpecificFarmKey{}).(string)
}

func calculateMaxResult(r *http.Request) (int, error) {
	maxResultPerpage := r.URL.Query().Get("max_result")
	if maxResultPerpage == "" {
		maxResultPerpage = "50"
	}

	maxResult, err := strconv.Atoi(maxResultPerpage)
	if err != nil {
		log.Error().Err(errors.Wrap(err, fmt.Sprintf("ERROR: invalid max result number %s", err))).Msg("")
		return 0, fmt.Errorf("error: invalid max result number : %w", err)
	}

	log.Info().Str("max result", fmt.Sprint(maxResult)).Msg("Preparing param max result")
	return maxResult, nil
}

func calculateOffset(maxResult int, r *http.Request) (int, error) {
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "0"
	}

	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		log.Error().Err(errors.Wrap(err, fmt.Sprintf("ERROR: invalid page number %s", err))).Msg("")
		return 0, fmt.Errorf("error: invalid page number : %w", err)
	}

	offset := 0
	if pageNumber > 1 {
		offset = pageNumber * maxResult
	}

	log.Info().Str("offset", fmt.Sprint(offset)).Msg("Preparing param page offset")
	return offset, nil
}

// HandleRequestsQueryParams takes the request and restore the query paramas, handle errors and set default values if not available
func (a *App) HandleRequestsQueryParams(r *http.Request) (*http.Request, error) {

	farmID := r.URL.Query().Get("farm_id")
	isSpecificFarm := ""
	if farmID != "" {
		isSpecificFarm = fmt.Sprintf(",where:{farmId_eq:%s}", farmID)
	} else {
		isSpecificFarm = ""
	}

	log.Info().Str("farm", fmt.Sprint(isSpecificFarm)).Msg("Preparing param specific farm id")

	maxResult, err := calculateMaxResult(r)
	if err != nil {
		return &http.Request{}, fmt.Errorf("error: invalid max result number : %w", err)
	}
	offset, err := calculateOffset(maxResult, r)
	if err != nil {
		return &http.Request{}, fmt.Errorf("error: invalid max result number : %w", err)
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, SpecificFarmKey{}, isSpecificFarm)
	ctx = context.WithValue(ctx, OffsetKey{}, offset)
	ctx = context.WithValue(ctx, MaxResultKey{}, maxResult)

	return r.WithContext(ctx), nil
}
