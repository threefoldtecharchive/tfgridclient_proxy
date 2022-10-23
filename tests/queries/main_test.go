package main

import (
	"flag"
	"math/rand"
	"os"
	"testing"

	// used by the orm

	_ "github.com/lib/pq"
)

func parseCmdline() {
	flag.StringVar(&POSTGRES_HOST, "postgres-host", "", "postgres host")
	flag.IntVar(&POSTGRES_PORT, "postgres-port", 5432, "postgres port")
	flag.StringVar(&POSTGRES_DB, "postgres-db", "", "postgres database")
	flag.StringVar(&POSTGRES_USER, "postgres-user", "", "postgres username")
	flag.StringVar(&POSTGRES_PASSSWORD, "postgres-password", "", "postgres password")
	flag.StringVar(&ENDPOINT, "endpoint", "", "the grid proxy endpoint to test against")
	flag.IntVar(&SEED, "seed", 0, "seed used for the random generation of tests")
	flag.Parse()
}

func TestMain(m *testing.M) {
	parseCmdline()
	if SEED != 0 {
		rand.Seed(int64(SEED))
	}

	exitcode := m.Run()
	os.Exit(exitcode)
}
