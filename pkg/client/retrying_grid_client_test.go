package client

import (
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/threefoldtech/grid_proxy_server/pkg/types"
)

type requestCounter struct {
	Counter int
}

func NewRequestCounter() Client {
	return &requestCounter{0}
}

func (r *requestCounter) Ping() error {
	r.Counter++
	return errors.New("error")
}
func (r *requestCounter) Nodes(filter types.NodeFilter, pagination types.Limit) (res []types.Node, totalCount int, err error) {
	r.Counter++
	return nil, 0, errors.New("error")
}
func (r *requestCounter) Farms(filter types.FarmFilter, pagination types.Limit) (res []types.Farm, totalCount int, err error) {
	r.Counter++
	return nil, 0, errors.New("error")
}
func (r *requestCounter) Contracts(filter types.ContractFilter, pagination types.Limit) (res []types.Contract, totalCount int, err error) {
	r.Counter++
	return nil, 0, errors.New("error")
}
func (r *requestCounter) Twins(filter types.TwinFilter, pagination types.Limit) (res []types.Twin, totalCount int, err error) {
	r.Counter++
	return nil, 0, errors.New("error")
}
func (r *requestCounter) Node(nodeID uint32) (res types.NodeWithNestedCapacity, err error) {
	r.Counter++
	return types.NodeWithNestedCapacity{}, errors.New("error")
}
func (r *requestCounter) NodeStatus(nodeID uint32) (res types.NodeStatus, err error) {
	r.Counter++
	return types.NodeStatus{}, errors.New("error")
}
func (r *requestCounter) Counters(filter types.StatsFilter) (res types.Counters, err error) {
	r.Counter++
	return types.Counters{}, errors.New("error")
}

func retryingConstructor(u string) Client {
	return NewRetryingClientWithTimeout(NewClient(u), 1*time.Millisecond)
}

func TestRetryingConnectionFailures(t *testing.T) {
	testConnectionFailures(t, retryingConstructor)
}

func TestRetryingPingFailure(t *testing.T) {
	testPingFailure(t, retryingConstructor)
}

func TestRetryingStatusCodeFailures(t *testing.T) {
	testStatusCodeFailures(t, retryingConstructor)
}

func TestRetryingSuccess(t *testing.T) {
	testSuccess(t, retryingConstructor)
}

func TestCalledMultipleTimes(t *testing.T) {
	r := NewRequestCounter()
	proxy := NewRetryingClientWithTimeout(r, 1*time.Millisecond)
	methods := map[string]func(){
		"nodes": func() {
			_, _, _ = proxy.Nodes(types.NodeFilter{}, types.Limit{})
		},
		"node": func() {
			_, _ = proxy.Node(1)
		},
		"farms": func() {
			_, _, _ = proxy.Farms(types.FarmFilter{}, types.Limit{})
		},
		"node_status": func() {
			_, _ = proxy.NodeStatus(1)
		},
	}
	for endpoint, f := range methods {
		beforeCount := r.(*requestCounter).Counter
		f()
		afterCount := r.(*requestCounter).Counter
		fmt.Printf("%d %d ", beforeCount, afterCount)
		if afterCount-beforeCount <= 1 {
			t.Fatalf("retrying %s client is expected to try more than once. before calls: %d, after calls: %d", endpoint, beforeCount, afterCount)
		}
	}
}
