package explorer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/zos/pkg/gridtypes"
	"github.com/threefoldtech/zos/pkg/rmb"
)

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
