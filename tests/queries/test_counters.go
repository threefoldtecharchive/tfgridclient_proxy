package main

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	proxyclient "github.com/threefoldtech/grid_proxy_server/pkg/client"
	proxytypes "github.com/threefoldtech/grid_proxy_server/pkg/types"
)

func validateCountersResults(local, remote proxytypes.Counters) error {
	if !reflect.DeepEqual(local, remote) {
		return fmt.Errorf("counters mismatch: local: %+v, remote: %+v", local, remote)
	}
	return nil
}

func countersAllTest(proxyClient, localClient proxyclient.Client) error {
	f := proxytypes.StatsFilter{}
	counters, err := localClient.Counters(f)
	if err != nil {
		return err
	}
	remote, err := proxyClient.Counters(f)
	if err != nil {
		return err
	}
	if err := validateCountersResults(counters, remote); err != nil {
		return errors.Wrapf(err, "filter: all")
	}
	return nil
}

func countersUpTest(proxyClient, localClient proxyclient.Client) error {
	f := proxytypes.StatsFilter{
		Status: &statusUP,
	}
	counters, err := localClient.Counters(f)
	if err != nil {
		return err
	}
	remote, err := proxyClient.Counters(f)
	if err != nil {
		return err
	}
	if err := validateCountersResults(counters, remote); err != nil {
		return errors.Wrapf(err, "filter: up")
	}
	return nil
}

func countersTest(proxyClient, localClient proxyclient.Client) error {
	if err := countersUpTest(proxyClient, localClient); err != nil {
		return err
	}
	if err := countersAllTest(proxyClient, localClient); err != nil {
		return err
	}
	return nil
}
