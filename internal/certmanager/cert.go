package certmanager

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// CertificateConfig the config the user passes to the certificate manager
type CertificateConfig struct {
	Domain   string
	Email    string
	CA       string
	CacheDir string
}

// CertificateData the result of the generated certificate
type CertificateData struct {
	KeyPath  string
	CertPath string
	Fresh    bool
}

// CertificateManager manages certificate generation
type CertificateManager struct {
	config   CertificateConfig
	provider *Provider
}

// NewCertificateManager creates a new certificate manager with the given config
func NewCertificateManager(config CertificateConfig) *CertificateManager {
	provider := &Provider{
		tokenAuths: make(map[string]string),
	}
	return &CertificateManager{
		config:   config,
		provider: provider,
	}
}

type user struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *user) GetEmail() string {
	return u.Email
}
func (u user) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *user) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func readTLSCert(cert []byte, key []byte) (*x509.Certificate, error) {

	tlsCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	crt := tlsCert.Leaf
	if crt == nil {
		crt, err = x509.ParseCertificate(tlsCert.Certificate[0])
		if err != nil {
			return nil, err
		}
	}
	return crt, nil
}

// EnsureCertificate checks the current certificate's expiry, and generates
//
//	a new one if no cert is found or its expiry date is less
//	than 30 day from now. And returns the certificate data.
func (c *CertificateManager) EnsureCertificate() (CertificateData, error) {
	certPath := filepath.Join(c.config.CacheDir, "cert.pem")
	keyPath := filepath.Join(c.config.CacheDir, "key.pem")
	newCert := false

	err := os.MkdirAll(c.config.CacheDir, 0644)
	if err != nil {
		return CertificateData{}, errors.Wrap(err, "couldn't create cache dir")
	}
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		newCert = true
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		newCert = true
	}
	key, err := os.ReadFile(keyPath)
	if err != nil {
		log.Warn().Err(err).Msg("couldn't read key from disk")
		newCert = true
	}

	cert, err := os.ReadFile(certPath)
	if err != nil {
		log.Warn().Err(err).Msg("couldn't read key from disk")
		newCert = true
	}
	if !newCert {
		crt, err := readTLSCert(cert, key)
		if err != nil {
			log.Warn().Err(err).Msg("couldn't read old tls certificate")
			newCert = true
		}
		err = crt.VerifyHostname(c.config.Domain)
		if err != nil {
			log.Warn().Err(err).Msg("an old certificate found but not containing the required domain")
			newCert = true
		}
		if !newCert && !crt.NotAfter.Before(time.Now().Add(24*30*time.Hour)) {
			log.Debug().Msg("found an old certificate with late-enough expiry date")
			return CertificateData{
				KeyPath:  keyPath,
				CertPath: certPath,
				Fresh:    false,
			}, nil
		}
	}
	// Create a user. New accounts need an email and private key to start.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return CertificateData{}, errors.Wrap(err, "couldn't generate key")
	}

	myUser := user{
		Email: c.config.Email,
		key:   privateKey,
	}

	config := lego.NewConfig(&myUser)

	// This CA URL is configured for a local dev instance of Boulder running in Docker in a VM.
	config.CADirURL = c.config.CA
	// config.CADirURL = "https://acme-staging-v02.api.letsencrypt.org/directory"
	config.Certificate.KeyType = certcrypto.RSA2048

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		return CertificateData{}, errors.Wrap(err, "couldn't get new lego client")
	}

	// We specify an HTTP port of 5002 and an TLS port of 5001 on all interfaces
	// because we aren't running as root and can't bind a listener to port 80 and 443
	// (used later when we attempt to pass challenges). Keep in mind that you still
	// need to proxy challenge traffic to port 5002 and 5001.
	err = client.Challenge.SetHTTP01Provider(c.provider)
	if err != nil {
		return CertificateData{}, errors.Wrap(err, "couldn't listen on http port")
	}

	// New users will need to register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return CertificateData{}, errors.Wrap(err, "couldn't register new account")
	}
	myUser.Registration = reg
	var certificates *certificate.Resource
	if newCert {
		request := certificate.ObtainRequest{
			Domains: []string{c.config.Domain},
			Bundle:  true,
		}
		certificates, err = client.Certificate.Obtain(request)
		if err != nil {
			return CertificateData{}, errors.Wrap(err, "couldn't obtain certificate")
		}
	} else {
		certificates, err = client.Certificate.Renew(certificate.Resource{
			Domain:      c.config.Domain,
			PrivateKey:  key,
			Certificate: cert,
		}, true, false, "")
		if err != nil {
			return CertificateData{}, errors.Wrap(err, "couldn't renew certificate")
		}
	}
	err = os.WriteFile(certPath, certificates.Certificate, 0644)
	if err != nil {
		return CertificateData{}, errors.Wrap(err, "couldn't write cert to disk")
	}
	err = os.WriteFile(keyPath, certificates.PrivateKey, 0644)
	if err != nil {
		return CertificateData{}, errors.Wrap(err, "couldn't write key to disk")
	}
	return CertificateData{
		KeyPath:  keyPath,
		CertPath: certPath,
		Fresh:    true,
	}, nil
	// ... all done.
}

// KeypairReloader to use with TLS servers for dynamic reloading
type KeypairReloader struct {
	certMu      sync.RWMutex
	cert        *tls.Certificate
	certManager *CertificateManager
}

// NewKeypairReloader creates a new instance given the manager to create certs
func NewKeypairReloader(certManager *CertificateManager) (*KeypairReloader, error) {
	result := &KeypairReloader{
		certManager: certManager,
	}
	certData, err := result.certManager.EnsureCertificate()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't update certificate")
	}
	_ = result.reload(certData.CertPath, certData.KeyPath)
	go func() {
		for range time.Tick(time.Hour * 24 * 5) {
			certData, err := result.certManager.EnsureCertificate()
			if err != nil {
				log.Error().Err(err).Msg("couldn't update certificate")
			}
			if certData.Fresh {
				err = result.reload(certData.CertPath, certData.KeyPath)
				if err != nil {
					log.Error().Err(err).Msg("failed to load newly created certificate")
				}
			}
		}
	}()
	return result, nil
}

func (kpr *KeypairReloader) reload(certPath, keyPath string) error {
	newCert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return err
	}
	kpr.certMu.Lock()
	defer kpr.certMu.Unlock()
	kpr.cert = &newCert
	return nil
}

// GetCertificateFunc returns a function that returns the up-to-date certificate
func (kpr *KeypairReloader) GetCertificateFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		kpr.certMu.RLock()
		defer kpr.certMu.RUnlock()
		return kpr.cert, nil
	}
}

// Provider to use for presenting tokens when generating certs
type Provider struct {
	// token -> authorization text
	tokenAuths map[string]string
}

// Present associates the token with keyAuth
func (p *Provider) Present(domain, token, keyAuth string) error {
	p.tokenAuths[token] = keyAuth
	return nil
}

// CleanUp removes the token entry
func (p *Provider) CleanUp(domain, token, keyAuth string) error {
	delete(p.tokenAuths, token)
	return nil
}

func (p *Provider) handler(w http.ResponseWriter, req *http.Request) {
	challengePrefix := "/.well-known/acme-challenge/"
	path := req.URL.Path
	log.Debug().Str("path", req.URL.Path).Msg("received a request")
	if strings.HasPrefix(path, challengePrefix) {
		token := strings.TrimPrefix(path, challengePrefix)
		auth, ok := p.tokenAuths[token]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(auth))
		}
	} else {
		http.Redirect(w, req, "https://"+req.Host+req.RequestURI, http.StatusMovedPermanently)
	}
}

// ListenForChallenges handles http cert verification requests and redirects
//
//	all other requests to https
func (c CertificateManager) ListenForChallenges() error {
	log.Info().Msg("Creating server")
	router := mux.NewRouter().StrictSlash(true)
	router.PathPrefix("/").Handler(http.HandlerFunc(c.provider.handler))

	s := &http.Server{
		Handler: router,
		Addr:    ":80",
	}
	if err := s.ListenAndServe(); err != nil {
		if err == http.ErrServerClosed {
			log.Info().Msg("server stopped gracefully")
			return nil
		}
		return err

	}
	return nil
}
