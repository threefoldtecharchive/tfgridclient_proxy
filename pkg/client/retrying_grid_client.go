package client

import (
	"log"
	"time"

	"github.com/cenkalti/backoff/v3"
	"github.com/threefoldtech/grid_proxy_server/pkg/types"
)

// RetryingClient wraps the given client and does the actions with retrying
type RetryingClient struct {
	cl      Client
	timeout time.Duration
}

// NewRetryingClient retrying grid proxy client constructor
func NewRetryingClient(cl Client) Client {
	return NewRetryingClientWithTimeout(cl, 2*time.Minute)
}

// NewRetryingClient retrying grid proxy client constructor with a timeout as a parameter
func NewRetryingClientWithTimeout(cl Client, timeout time.Duration) Client {
	proxy := RetryingClient{cl, timeout}
	return &proxy
}

func bf(timeout time.Duration) *backoff.ExponentialBackOff {
	res := backoff.NewExponentialBackOff()
	res.MaxElapsedTime = timeout
	return res
}

func notify(cmd string) func(error, time.Duration) {
	return func(err error, duration time.Duration) {
		log.Printf("failure: %s, command: %s, duration: %s", err.Error(), cmd, duration)
	}
}

// Ping makes sure the server is up
func (g *RetryingClient) Ping() error {
	f := func() error {
		return g.cl.Ping()
	}
	return backoff.RetryNotify(f, bf(g.timeout), notify("ping"))

}

// Nodes returns nodes with the given filters and pagination parameters
func (g *RetryingClient) Nodes(filter types.NodeFilter, pagination types.Limit) (res []types.Node, err error) {
	f := func() error {
		res, err = g.cl.Nodes(filter, pagination)
		return err
	}
	err = backoff.RetryNotify(f, bf(g.timeout), notify("nodes"))
	return
}

// Farms returns farms with the given filters and pagination parameters
func (g *RetryingClient) Farms(filter types.FarmFilter, pagination types.Limit) (res []types.Farm, err error) {
	f := func() error {
		res, err = g.cl.Farms(filter, pagination)
		return err
	}
	err = backoff.RetryNotify(f, bf(g.timeout), notify("farms"))
	return
}

// Contracts returns contracts with the given filters and pagination parameters
func (g *RetryingClient) Contracts(filter types.ContractFilter, pagination types.Limit) (res []types.Contract, err error) {
	f := func() error {
		res, err = g.cl.Contracts(filter, pagination)
		return err
	}
	err = backoff.RetryNotify(f, bf(g.timeout), notify("contracts"))
	return
}

// Node returns the node with the give id
func (g *RetryingClient) Node(nodeID uint32) (res types.NodeWithNestedCapacity, err error) {
	f := func() error {
		res, err = g.cl.Node(nodeID)
		return err
	}
	err = backoff.RetryNotify(f, bf(g.timeout), notify("node"))
	return
}

// Node returns the node with the give id
func (g *RetryingClient) NodeStatus(nodeID uint32) (res types.NodeStatus, err error) {
	f := func() error {
		res, err = g.cl.NodeStatus(nodeID)
		return err
	}
	err = backoff.RetryNotify(f, bf(g.timeout), notify("node_status"))
	return
}
