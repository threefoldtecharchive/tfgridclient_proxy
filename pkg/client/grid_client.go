package client

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"github.com/threefoldtech/grid_proxy_server/pkg/types"
)

// Client a client to communicate with the grid proxy
type Client interface {
	Ping() error
	Nodes(filter types.NodeFilter, pagination types.Limit) (res []types.Node, err error)
	Farms(filter types.FarmFilter, pagination types.Limit) (res []types.Farm, err error)
	Contracts(filter types.ContractFilter, pagination types.Limit) (res []types.Contract, err error)
	Node(nodeID uint32) (res types.NodeWithNestedCapacity, err error)
	NodeStatus(nodeID uint32) (res types.NodeStatus, err error)
}

// Clientimpl concrete implementation of the client to communicate with the grid proxy
type Clientimpl struct {
	endpoint string
}

// NewClient grid proxy client constructor
func NewClient(endpoint string) Client {
	if endpoint[len(endpoint)-1] != '/' {
		endpoint += "/"
	}
	proxy := Clientimpl{endpoint}
	return &proxy
}

func parseError(body io.ReadCloser) error {
	text, err := ioutil.ReadAll(body)
	if err != nil {
		return errors.Wrap(err, "couldn't read body response")
	}
	var res ErrorReply
	if err := json.Unmarshal(text, &res); err != nil {
		return errors.New(string(text))
	}
	return fmt.Errorf("%s", res.Error)
}

func (g *Clientimpl) url(sub string, args ...interface{}) string {
	return g.endpoint + fmt.Sprintf(sub, args...)
}

// Ping makes sure the server is up
func (g *Clientimpl) Ping() error {
	req, err := http.Get(g.url(""))
	if err != nil {
		return err
	}
	if req.StatusCode != http.StatusOK {
		return fmt.Errorf("non ok return status code from the the grid proxy home page: %s", http.StatusText(req.StatusCode))
	}
	return nil
}

// Nodes returns nodes with the given filters and pagination parameters
func (g *Clientimpl) Nodes(filter types.NodeFilter, limit types.Limit) (res []types.Node, err error) {
	query := nodeParams(filter, limit)
	req, err := http.Get(g.url(fmt.Sprintf("nodes%s", query)))
	if err != nil {
		return
	}
	if req.StatusCode != http.StatusOK {
		err = parseError(req.Body)
		return
	}
	if err := json.NewDecoder(req.Body).Decode(&res); err != nil {
		return res, err
	}
	return
}

// Farms returns farms with the given filters and pagination parameters
func (g *Clientimpl) Farms(filter types.FarmFilter, limit types.Limit) (res []types.Farm, err error) {
	query := farmParams(filter, limit)
	req, err := http.Get(g.url(fmt.Sprintf("farms%s", query)))
	if err != nil {
		return
	}
	if req.StatusCode != http.StatusOK {
		err = parseError(req.Body)
		return
	}
	data, err := io.ReadAll(req.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &res)
	return
}

// Contracts returns contracts with the given filters and pagination parameters
func (g *Clientimpl) Contracts(filter types.ContractFilter, limit types.Limit) (res []types.Contract, err error) {
	query := contractParams(filter, limit)
	req, err := http.Get(g.url(fmt.Sprintf("contracts%s", query)))
	if err != nil {
		return
	}
	if req.StatusCode != http.StatusOK {
		err = parseError(req.Body)
		return
	}
	data, err := io.ReadAll(req.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return
	}
	for idx := range res {
		if res[idx].Type == "node" {
			res[idx].Details = types.NodeContractDetails{
				NodeID:            uint(res[idx].Details.(map[string]interface{})["nodeId"].(float64)),
				DeploymentData:    res[idx].Details.(map[string]interface{})["deployment_data"].(string),
				DeploymentHash:    res[idx].Details.(map[string]interface{})["deployment_hash"].(string),
				NumberOfPublicIps: uint(res[idx].Details.(map[string]interface{})["number_of_public_ips"].(float64)),
			}
		} else if res[idx].Type == "rent" {
			res[idx].Details = types.RentContractDetails{
				NodeID: uint(res[idx].Details.(map[string]interface{})["nodeId"].(float64)),
			}
		} else if res[idx].Type == "name" {
			res[idx].Details = types.NameContractDetails{
				Name: res[idx].Details.(map[string]interface{})["name"].(string),
			}
		}
	}
	return
}

// Node returns the node with the give id
func (g *Clientimpl) Node(nodeID uint32) (res types.NodeWithNestedCapacity, err error) {
	req, err := http.Get(g.url("nodes/%d", nodeID))
	if err != nil {
		return
	}
	if req.StatusCode != http.StatusOK {
		err = parseError(req.Body)
		return
	}
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &res)
	return
}

// Node returns the node status up/down
func (g *Clientimpl) NodeStatus(nodeID uint32) (res types.NodeStatus, err error) {
	req, err := http.Get(g.url("nodes/%d/status", nodeID))
	if err != nil {
		return
	}
	if req.StatusCode != http.StatusOK {
		err = parseError(req.Body)
		return
	}
	if err := json.NewDecoder(req.Body).Decode(&res); err != nil {
		return res, err
	}
	return
}
