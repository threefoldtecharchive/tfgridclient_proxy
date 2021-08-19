package explorer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/zos/pkg/gridtypes"
	"github.com/threefoldtech/zos/pkg/rmb"
)

func NewNodeClient(nodeTwin uint32, bus rmb.Client) *NodeClient {
	return &NodeClient{nodeTwin, bus}
}

func getNodeTwinId(nodeId string) uint32 {

	queryString := fmt.Sprintf(`
	{
		nodes(limit:10, where:{nodeId_eq:%s}){
		  twinId
		}
	}
	`, nodeId)

	result := []byte(query(queryString))

	var res NodeResult
	err := json.Unmarshal(result, &res)

	if err != nil {
		log.Error().Err(errors.Wrap(err, "couldn't parse json")).Msg("connection error")
	}
	nodeStats := res.Data.NodeResult
	if len(nodeStats) > 0 {
		return nodeStats[0].TwinId
	} else {
		return 0
	}

}

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

func query(jsonQuery string) string {
	jsonData := map[string]string{
		"query": jsonQuery,
	}
	jsonValue, _ := json.Marshal(jsonData)
	request, err := http.NewRequest("POST", "https://explorer.devnet.grid.tf/graphql/", bytes.NewBuffer(jsonValue))
	request.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Error().Err(errors.Wrap(err, "Failed to connect to graphql network")).Msg("connection error")
	}

	client := &http.Client{Timeout: time.Second * 10}
	response, err := client.Do(request)
	if err != nil {
		log.Error().Err(errors.Wrap(err, "Failed to connect to graphql network")).Msg("connection error")
	}
	defer response.Body.Close()
	if err != nil {
		log.Error().Err(errors.Wrap(err, "Failed to connect to graphql network, Response Failed")).Msg("connection error")
	}
	data, _ := ioutil.ReadAll(response.Body)
	return string(data)
}
