//go:build integration
// +build integration

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	proxyclient "github.com/threefoldtech/grid_proxy_server/pkg/client"
	proxytypes "github.com/threefoldtech/grid_proxy_server/pkg/types"
)

var postgresHost = flag.String("postgres-host", "", "postgres host")
var postgresPort = flag.Int("postgres-port", 5432, "postgres port")
var postgresDB = flag.String("postgres-db", "", "postgres database")
var postgresUser = flag.String("postgres-user", "", "postgres username")
var postgresPassword = flag.String("postgres-password", "", "postgres password")
var endpoint = flag.String("endpoint", "", "the grid proxy endpoint to test against")
var seed = flag.Int("seed", 0, "seed used for the random generation of tests")

var data DBData
var proxyClient, localClient proxyclient.Client

func TestMain(m *testing.M) {
	flag.Parse()
	if *seed != 0 {
		rand.Seed(int64(*seed))
	}
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		*postgresHost, *postgresPort, *postgresUser, *postgresPassword, *postgresDB)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(errors.Wrap(err, "failed to open db"))
	}
	defer db.Close()
	data, err = load(db)
	if err != nil {
		panic(err)
	}
	proxyClient = proxyclient.NewClient(*endpoint)
	localClient = NewGridProxyClient(data)
	os.Exit(m.Run())
}

func TestCountersUp(t *testing.T) {
	f := proxytypes.StatsFilter{
		Status: &statusUP,
	}
	counters, err := localClient.Counters(f)
	assert.NoError(t, err)
	remote, err := proxyClient.Counters(f)
	assert.NoError(t, err)
	assert.Equal(t, counters, remote)
}

func TestCountersAll(t *testing.T) {
	f := proxytypes.StatsFilter{}
	counters, err := localClient.Counters(f)
	assert.NoError(t, err)
	remote, err := proxyClient.Counters(f)
	assert.NoError(t, err)
	assert.Equal(t, counters, remote)
}
