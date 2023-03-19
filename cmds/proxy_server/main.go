package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/grid_proxy_server/internal/certmanager"
	"github.com/threefoldtech/grid_proxy_server/internal/explorer"
	"github.com/threefoldtech/grid_proxy_server/internal/explorer/db"
	logging "github.com/threefoldtech/grid_proxy_server/pkg"
	"github.com/threefoldtech/rmb-sdk-go"
	"github.com/threefoldtech/rmb-sdk-go/direct"
	"github.com/threefoldtech/substrate-client"
)

const (
	// CertDefaultCacheDir directory to keep the genreated certificates
	CertDefaultCacheDir = "/tmp/certs"
	DefaultTFChainURL   = "wss://tfchain.dev.grid.tf/ws"
	DefaultRelayURL     = "wss://relay.dev.grid.tf"
)

// GitCommit holds the commit version
var GitCommit string

type flags struct {
	debug            string
	postgresHost     string
	postgresPort     int
	postgresDB       string
	postgresUser     string
	postgresPassword string
	address          string
	version          bool
	nocert           bool
	domain           string
	TLSEmail         string
	CA               string
	certCacheDir     string
	tfChainURL       string
	relayURL         string
	mnemonics        string
}

func main() {
	f := flags{}
	flag.StringVar(&f.debug, "log-level", "info", "log level [debug|info|warn|error|fatal|panic]")
	flag.StringVar(&f.address, "address", ":443", "explorer running ip address")
	flag.StringVar(&f.postgresHost, "postgres-host", "", "postgres host")
	flag.IntVar(&f.postgresPort, "postgres-port", 5432, "postgres port")
	flag.StringVar(&f.postgresDB, "postgres-db", "", "postgres database")
	flag.StringVar(&f.postgresUser, "postgres-user", "", "postgres username")
	flag.StringVar(&f.postgresPassword, "postgres-password", "", "postgres password")
	flag.BoolVar(&f.version, "v", false, "shows the package version")
	flag.BoolVar(&f.nocert, "no-cert", false, "start the server without certificate")
	flag.StringVar(&f.domain, "domain", "", "domain on which the server will be served")
	flag.StringVar(&f.TLSEmail, "email", "", "tmail address to generate certificate with")
	flag.StringVar(&f.CA, "ca", "https://acme-v02.api.letsencrypt.org/directory", "certificate authority used to generate certificate")
	flag.StringVar(&f.certCacheDir, "cert-cache-dir", CertDefaultCacheDir, "path to store generated certs in")
	flag.StringVar(&f.tfChainURL, "tfchain-url", DefaultTFChainURL, "TF chain url")
	flag.StringVar(&f.relayURL, "relay-url", DefaultRelayURL, "RMB relay url")
	flag.StringVar(&f.mnemonics, "mnemonics", "", "Dummy user mnemonics for relay calls")
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
	if f.mnemonics == "" {
		log.Fatal().Msg("mnemonics are required")
	}
	logging.SetupLogging(f.debug)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	subManager := substrate.NewManager(f.tfChainURL)
	sub, err := subManager.Substrate()
	if err != nil {
		log.Fatal().Err(err).Msg(fmt.Sprintf("failed to connect to TF chain URL: %s", err))
	}
	defer sub.Close()

	relayClient, err := createRMBClient(ctx, f.relayURL, f.mnemonics, sub)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create realy client")
	}

	s, err := createServer(f, GitCommit, relayClient)
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

	config := certmanager.CertificateConfig{
		Domain:   f.domain,
		Email:    f.TLSEmail,
		CA:       f.CA,
		CacheDir: f.certCacheDir,
	}
	cm := certmanager.NewCertificateManager(config)
	go func() {
		if err := cm.ListenForChallenges(); err != nil {
			log.Error().Err(err).Msg("error occurred when listening for challenges")
		}
	}()
	kpr, err := certmanager.NewKeypairReloader(cm)
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

func createRMBClient(ctx context.Context, relayURL, mnemonics string, sub *substrate.Substrate) (rmb.Client, error) {
	client, err := direct.NewClient(ctx, direct.KeyTypeSr25519, mnemonics, relayURL, "tfgrid_proxy", sub, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create direct RMB client: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping Relay server: %s", err)
	}
	return client, nil
}

func createServer(f flags, gitCommit string, relayClient rmb.Client) (*http.Server, error) {
	log.Info().Msg("Creating server")

	router := mux.NewRouter().StrictSlash(true)
	db, err := db.NewPostgresDatabase(f.postgresHost, f.postgresPort, f.postgresUser, f.postgresPassword, f.postgresDB)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get postgres client")
	}

	// setup explorer
	if err := explorer.Setup(router, gitCommit, db, relayClient); err != nil {
		return nil, err
	}

	return &http.Server{
		Handler: router,
		Addr:    f.address,
	}, nil
}
