package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/grid_proxy_server/internal/explorer"
	"github.com/threefoldtech/grid_proxy_server/internal/explorer/db"
	"github.com/threefoldtech/grid_proxy_server/internal/rmbproxy"
	logging "github.com/threefoldtech/grid_proxy_server/pkg"
	"github.com/threefoldtech/substrate-client"
	"github.com/threefoldtech/zos/pkg/rmb"
)

const (
	// CertDefaultCacheDir directory to keep the genreated certificates
	CertDefaultCacheDir = "/tmp/certs"
)

// GitCommit holds the commit version
var GitCommit string

type flags struct {
	debug            string
	redis            string
	postgresHost     string
	postgresPort     int
	postgresDB       string
	postgresUser     string
	postgresPassword string
	address          string
	substrate        string
	domain           string
	TLSEmail         string
	CA               string
	certCacheDir     string
	version          bool
	nocert           bool
}

type api struct {
	version   string
	router    *mux.Router
	rmbClient rmb.Client
	c         *cache.Cache
	gitCommit string
	database  db.Database
	resolver  *rmbproxy.TwinExplorerResolver
}

func main() {
	f := flags{}
	flag.StringVar(&f.debug, "log-level", "info", "log level [debug|info|warn|error|fatal|panic]")
	flag.StringVar(&f.substrate, "substrate", "wss://tfchain.dev.grid.tf/ws", "substrate url")
	flag.StringVar(&f.address, "address", ":443", "explorer running ip address")
	flag.StringVar(&f.domain, "domain", "", "domain on which the server will be served")
	flag.StringVar(&f.TLSEmail, "email", "", "tmail address to generate certificate with")
	flag.StringVar(&f.CA, "ca", "https://acme-staging-v02.api.letsencrypt.org/directory", "certificate authority used to generate certificate")
	flag.StringVar(&f.postgresHost, "postgres-host", "", "postgres host")
	flag.IntVar(&f.postgresPort, "postgres-port", 5432, "postgres port")
	flag.StringVar(&f.postgresDB, "postgres-db", "", "postgres database")
	flag.StringVar(&f.postgresUser, "postgres-user", "", "postgres username")
	flag.StringVar(&f.postgresPassword, "postgres-password", "", "postgres password")
	flag.StringVar(&f.redis, "redis", "tcp://127.0.0.1:6379", "redis url")
	flag.BoolVar(&f.version, "v", false, "shows the package version")
	flag.StringVar(&f.certCacheDir, "cert-cache-dir", CertDefaultCacheDir, "path to store generated certs in")
	flag.BoolVar(&f.nocert, "no-cert", false, "start the server without certificate")
	flag.Parse()

	// shows version and exit
	if f.version {
		fmt.Printf("git rev: %s\n", GitCommit)
		os.Exit(0)
	}

	if f.domain == "" {
		log.Fatal().Err(errors.New("domain is required"))
	}
	if f.TLSEmail == "" {
		log.Fatal().Err(errors.New("email is required"))
	}

	logging.SetupLogging(f.debug)
	substrate, err := substrate.NewManager(f.substrate).Substrate()
	if err != nil {
		log.Fatal().Err(errors.Wrap(err, "error in connecting to substrate"))
	}
	defer substrate.Close()
	s, err := createServer(f, GitCommit, substrate)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create mux server")
	}

	if err := app(s, f); err != nil {
		log.Fatal().Msg(err.Error())
	}

}

func app(s *http.Server, f flags) error {

	if f.nocert {
		log.Info().Str("listening on", f.address).Msg("Server started ...")
		if err := s.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				log.Info().Msg("server stopped gracefully")
			} else {
				log.Error().Err(err).Msg("server stopped unexpectedly")
			}
		}
		return nil
	}

	config := rmbproxy.CertificateConfig{
		Domain:   f.domain,
		Email:    f.TLSEmail,
		CA:       f.CA,
		CacheDir: f.certCacheDir,
	}
	cm := rmbproxy.NewCertificateManager(config)
	go func() {
		if err := cm.ListenForChallenges(); err != nil {
			log.Error().Err(err).Msg("error occurred when listening for challenges")
		}
	}()
	kpr, err := rmbproxy.NewKeypairReloader(cm)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initiate key reloader")
	}
	s.TLSConfig = &tls.Config{
		GetCertificate: kpr.GetCertificateFunc(),
	}

	log.Info().Str("listening on", f.address).Msg("Server started ...")
	if err := s.ListenAndServeTLS("", ""); err != nil {
		if err == http.ErrServerClosed {
			log.Info().Msg("server stopped gracefully")
		} else {
			log.Error().Err(err).Msg("server stopped unexpectedly")
		}
	}
	return nil
}

func createServer(f flags, gitCommit string, substrate *substrate.Substrate) (*http.Server, error) {
	// main router
	log.Info().Msg("Starting server")
	router := mux.NewRouter().StrictSlash(true)

	// postgres client
	log.Info().Msg("Preparing Postgres Client ...")
	db, err := db.NewPostgresDatabase(f.postgresHost, f.postgresPort, f.postgresUser, f.postgresPassword, f.postgresDB)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get postgres client")
	}

	// redis pool
	log.Info().Str("redis address", f.redis).Msg("Preparing Redis Pool ...")
	rmbClient, err := rmb.NewClient(f.redis, 500)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't connect to rmb")
	}
	c := cache.New(2*time.Minute, 3*time.Minute)

	// twin resolver
	log.Info().Msg("Creating Twin resolver")

	resolver, err := rmbproxy.NewTwinResolver(substrate)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get a client to explorer resolver")
	}

	// setup API v1
	router1 := router.PathPrefix("").Subrouter()
	api1 := api{
		version:   "v1",
		router:    router1,
		rmbClient: rmbClient,
		c:         c,
		gitCommit: gitCommit,
		database:  db,
		resolver:  resolver,
	}
	if err := setup(api1); err != nil {
		return nil, err
	}

	// setup API v2
	router2 := router.PathPrefix("/api/v2").Subrouter()
	api2 := api{
		version:   "v2",
		router:    router2,
		rmbClient: rmbClient,
		c:         c,
		gitCommit: gitCommit,
		database:  db,
		resolver:  resolver,
	}
	if err := setup(api2); err != nil {
		return nil, err
	}

	return &http.Server{
		Handler: router,
		Addr:    f.address,
	}, nil
}

func setup(api api) error {
	if err := explorer.Setup(api.version, api.router, api.rmbClient, api.c, api.gitCommit, api.database); err != nil {
		return err
	}
	if err := rmbproxy.Setup(api.version, api.router, api.resolver); err != nil {
		return err
	}
	return nil
}
