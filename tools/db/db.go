package main

import (
	"database/sql"
	"flag"
	"fmt"
	"math/rand"

	// used by the orm

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type flags struct {
	postgresHost     string
	postgresPort     int
	postgresDB       string
	postgresUser     string
	postgresPassword string
	reset            bool
	seed             int
}

func parseCmdline() flags {
	f := flags{}
	flag.StringVar(&f.postgresHost, "postgres-host", "", "postgres host")
	flag.IntVar(&f.postgresPort, "postgres-port", 5432, "postgres port")
	flag.StringVar(&f.postgresDB, "postgres-db", "", "postgres database")
	flag.StringVar(&f.postgresUser, "postgres-user", "", "postgres username")
	flag.StringVar(&f.postgresPassword, "postgres-password", "", "postgres password")
	flag.BoolVar(&f.reset, "reset", false, "reset the db before starting")
	flag.IntVar(&f.seed, "seed", 0, "seed used for the random generation of tests")
	flag.Parse()
	return f
}

func main() {
	f := parseCmdline()
	if f.seed != 0 {
		rand.Seed(int64(f.seed))
	}
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		f.postgresHost, f.postgresPort, f.postgresUser, f.postgresPassword, f.postgresDB)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(errors.Wrap(err, "failed to open db"))
	}
	defer db.Close()
	if f.reset {
		if _, err := db.Exec(
			`
			DROP TABLE IF EXISTS account CASCADE;
			DROP TABLE IF EXISTS burn_transaction CASCADE;
			DROP TABLE IF EXISTS city CASCADE;
			DROP TABLE IF EXISTS contract_bill_report CASCADE;
			DROP TABLE IF EXISTS contract_resources CASCADE;
			DROP TABLE IF EXISTS country CASCADE;
			DROP TABLE IF EXISTS entity CASCADE;
			DROP TABLE IF EXISTS entity_proof CASCADE;
			DROP TABLE IF EXISTS farm CASCADE;
			DROP TABLE IF EXISTS farming_policy CASCADE;
			DROP TABLE IF EXISTS historical_balance CASCADE;
			DROP TABLE IF EXISTS interfaces CASCADE;
			DROP TABLE IF EXISTS location CASCADE;
			DROP TABLE IF EXISTS migrations CASCADE;
			DROP TABLE IF EXISTS mint_transaction CASCADE;
			DROP TABLE IF EXISTS name_contract CASCADE;
			DROP TABLE IF EXISTS node CASCADE;
			DROP TABLE IF EXISTS node_contract CASCADE;
			DROP TABLE IF EXISTS node_resources_free CASCADE;
			DROP TABLE IF EXISTS node_resources_total CASCADE;
			DROP TABLE IF EXISTS node_resources_used CASCADE;
			DROP TABLE IF EXISTS nru_consumption CASCADE;
			DROP TABLE IF EXISTS pricing_policy CASCADE;
			DROP TABLE IF EXISTS public_config CASCADE;
			DROP TABLE IF EXISTS public_ip CASCADE;
			DROP TABLE IF EXISTS refund_transaction CASCADE;
			DROP TABLE IF EXISTS rent_contract CASCADE;
			DROP TABLE IF EXISTS transfer CASCADE;
			DROP TABLE IF EXISTS twin CASCADE;
			DROP TABLE IF EXISTS typeorm_metadata CASCADE;
			DROP TABLE IF EXISTS uptime_event CASCADE;
			DROP SCHEMA IF EXISTS substrate_threefold_status CASCADE;
			
		`); err != nil {
			panic(err)
		}
	}
	if err := initSchema(db); err != nil {
		panic(err)
	}
	// it looks like a useless block but everything breaks when it's removed
	_, err = db.Query("SELECT current_database();")
	if err != nil {
		panic(err)
	}
	// ----
	if err := generateData(db); err != nil {
		panic(err)
	}
}
