package explorer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/threefoldtech/grid_proxy_server/internal/explorer/db"
)

const (
	nodeQueryTmpl = `
	{
		nodes(limit:%d,offset:%d){
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
	`
	farmQueryTmpl = `
	{
		farms (limit:%d,offset:%d) {
			name
			farmId
			twinId
			version
			farmId
			pricingPolicyId
			stellarAddress
			publicIPs{
				id
				ip
				contractId
				gateway
			}
		}
	}`
)

type GraphqlClient struct {
	explorer string
}

func NewGraphqLClient(explorer string) GraphqlClient {
	return GraphqlClient{explorer}
}

func (a *GraphqlClient) baseQuery(queryString string) (io.ReadCloser, error) {
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

func (a *GraphqlClient) query(queryString string, result interface{}) error {
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

type NodeCursor struct {
	client   *GraphqlClient
	current  int
	pageSize int
}

func NewNodeCursor(client *GraphqlClient, pageSize int) NodeCursor {
	return NodeCursor{client, 0, pageSize}
}

func (n *NodeCursor) Next() ([]db.GraphqlData, error) {
	queryString := fmt.Sprintf(nodeQueryTmpl, n.pageSize, n.current)
	var nodes nodesResponse
	err := n.client.query(queryString, &nodes)
	if err != nil {
		return nil, fmt.Errorf("failed to query nodes %w", err)
	}
	n.current += len(nodes.Nodes.Data)
	return nodes.Nodes.Data, nil
}

type FarmCursor struct {
	client   *GraphqlClient
	current  int
	pageSize int
}

func NewFarmCursor(client *GraphqlClient, pageSize int) FarmCursor {
	return FarmCursor{client, 0, pageSize}
}

func (n *FarmCursor) Next() ([]db.Farm, error) {
	queryString := fmt.Sprintf(farmQueryTmpl, n.pageSize, n.current)
	var farms FarmResult
	err := n.client.query(queryString, &farms)
	if err != nil {
		return nil, fmt.Errorf("failed to query farms %w", err)
	}
	n.current += len(farms.Data.Farms)
	return farms.Data.Farms, nil
}
