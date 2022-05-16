package main

import (
	"database/sql"
	"flag"
	"fmt"

	// used by the orm

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/threefoldtech/grid_proxy_server/pkg/gridproxy"
)

type flags struct {
	postgresHost     string
	postgresPort     int
	postgresDB       string
	postgresUser     string
	postgresPassword string
	endpoint         string
}

func parseCmdline() flags {
	f := flags{}
	flag.StringVar(&f.postgresHost, "postgres-host", "", "postgres host")
	flag.IntVar(&f.postgresPort, "postgres-port", 5432, "postgres port")
	flag.StringVar(&f.postgresDB, "postgres-db", "", "postgres database")
	flag.StringVar(&f.postgresUser, "postgres-user", "", "postgres username")
	flag.StringVar(&f.postgresPassword, "postgres-password", "", "postgres password")
	flag.StringVar(&f.endpoint, "endpoint", "", "the grid proxy endpoint to test against")
	flag.Parse()
	return f
}

func main() {
	f := parseCmdline()
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
	proxyClient := gridproxy.NewGridProxyClient(f.endpoint)
	localClient := NewGridProxyClient(data)
	if err := NodesTest(&data, proxyClient, localClient); err != nil {
		panic(err)
	}
	if err := FarmsTest(&data, proxyClient, localClient); err != nil {
		panic(err)
	}
	if err := ContractsTest(&data, proxyClient, localClient); err != nil {
		panic(err)
	}
}
