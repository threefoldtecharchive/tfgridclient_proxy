package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/threefoldtech/grid_proxy_server/internal/explorer"
	"github.com/threefoldtech/grid_proxy_server/internal/explorer/db"
	"github.com/threefoldtech/grid_proxy_server/internal/rmbproxy"
	logging "github.com/threefoldtech/grid_proxy_server/pkg"
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
	flag.StringVar(&f.redis, "redis", "127.0.0.1:6379", "redis url")
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
	s, err := createServer(f, GitCommit)
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

func createServer(f flags, gitCommit string) (*http.Server, error) {
	log.Info().Msg("Creating server")

	router := mux.NewRouter().StrictSlash(true)
	db, err := db.NewPostgresDatabase(f.postgresHost, f.postgresPort, f.postgresUser, f.postgresPassword, f.postgresDB)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get postgres client")
	}

	// setup explorer
	if err := explorer.Setup(router, gitCommit, db); err != nil {
		return nil, err
	}

	// setup rmb proxy
	redis := redis.NewClient(&redis.Options{
		Addr: f.redis,
	})
	ttl := 10 * time.Minute
	proxy, err := rmbproxy.NewProxy(f.substrate, redis, ttl)
	if err != nil {
		return nil, err
	}
	proxy.Setup(router)

	return &http.Server{
		Handler: router,
		Addr:    f.address,
	}, nil
}
