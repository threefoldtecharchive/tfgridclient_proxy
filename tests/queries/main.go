package main

import (
	"database/sql"
	"flag"
	"fmt"
	"math/rand"

	// used by the orm

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	proxyclient "github.com/threefoldtech/grid_proxy_server/pkg/client"
)

type flags struct {
	postgresHost     string
	postgresPort     int
	postgresDB       string
	postgresUser     string
	postgresPassword string
	endpoint         string
	seed             int
	twins            bool
	farms            bool
	nodes            bool
	contracts        bool
	counters         bool
}

func parseCmdline() flags {
	f := flags{}
	flag.StringVar(&f.postgresHost, "postgres-host", "", "postgres host")
	flag.IntVar(&f.postgresPort, "postgres-port", 5432, "postgres port")
	flag.StringVar(&f.postgresDB, "postgres-db", "", "postgres database")
	flag.StringVar(&f.postgresUser, "postgres-user", "", "postgres username")
	flag.StringVar(&f.postgresPassword, "postgres-password", "", "postgres password")
	flag.StringVar(&f.endpoint, "endpoint", "", "the grid proxy endpoint to test against")
	flag.IntVar(&f.seed, "seed", 0, "seed used for the random generation of tests")
	flag.BoolVar(&f.nodes, "nodes", false, "run nodes test (run by default if all selectors are omitted)")
	flag.BoolVar(&f.contracts, "contracts", false, "run contracts test (run by default if all selectors are omitted)")
	flag.BoolVar(&f.twins, "twins", false, "run twins test (run by default if all selectors are omitted)")
	flag.BoolVar(&f.farms, "farms", false, "run farms test (run by default if all selectors are omitted)")
	flag.BoolVar(&f.counters, "counters", false, "run counters test (run by default if all selectors are omitted)")
	flag.Parse()
	return f
}

func main() {
	f := parseCmdline()
	if f.seed != 0 {
		rand.Seed(int64(f.seed))
	}
	if !f.twins && !f.contracts && !f.nodes && !f.farms && !f.counters {
		f.twins = true
		f.nodes = true
		f.farms = true
		f.contracts = true
		f.counters = true
	}
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		f.postgresHost, f.postgresPort, f.postgresUser, f.postgresPassword, f.postgresDB)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(errors.Wrap(err, "failed to open db"))
	}
	defer db.Close()
	data, err := load(db)
	if err != nil {
		panic(err)
	}
	proxyClient := proxyclient.NewClient(f.endpoint)
	localClient := NewGridProxyClient(data)
	if f.nodes {
		if err := nodesTest(&data, proxyClient, localClient); err != nil {
			panic(err)
		}
	}
	if f.farms {
		if err := farmsTest(&data, proxyClient, localClient); err != nil {
			panic(err)
		}
	}
	if f.contracts {
		if err := contractsTest(&data, proxyClient, localClient); err != nil {
			panic(err)
		}
	}
	if f.twins {
		if err := twinsTest(&data, proxyClient, localClient); err != nil {
			panic(err)
		}
	}
	if f.counters {
		if err := countersTest(proxyClient, localClient); err != nil {
			panic(err)
		}
	}
}
