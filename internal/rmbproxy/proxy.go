package rmbproxy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/threefoldtech/grid_proxy_server/internal/explorer/mw"

	// swagger configuration
	"github.com/pkg/errors"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/threefoldtech/go-rmb"
)

// App is the main app objects
type Proxy struct {
	resolver *TwinResolver
	rmb      *Rmb
}

// NewTwinClient : create new TwinClient
func NewProxy(substrateURL string, redis *redis.Client, cacheTtl time.Duration) (*Proxy, error) {
	resolver, err := NewTwinResolver(substrateURL, redis, cacheTtl)
	if err != nil {
		return nil, err
	}

	rmb := NewRmb(redis, cacheTtl)
	return &Proxy{
		resolver: resolver,
		rmb:      rmb,
	}, nil
}

// sendMessage godoc
// @Summary submit the message
// @Description submit the message
// @Tags RMB
// @Accept  json
// @Produce  json
// @Param msg body Message true "rmb.Message"
// @Param twin_id path int true "twin id"
// @Success 200 {object} MessageIdentifier
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Failure 502 {object} string
// @Router /twin/{twin_id} [post]
func (p *Proxy) sendMessage(r *http.Request) (interface{}, mw.Response) {
	twinIdString := mux.Vars(r)["twin_id"]

	buffer := new(bytes.Buffer)
	if _, err := buffer.ReadFrom(r.Body); err != nil {
		return nil, mw.BadRequest(err)
	}

	twinId, err := strconv.Atoi(twinIdString)
	if err != nil {
		return nil, mw.BadRequest(errors.Wrap(err, "invalid twin_id"))
	}

	_, err = p.resolver.Get(twinId)
	if err != nil {
		log.Error().Err(err).Msg("failed to get twin")
		return nil, mw.Error(errors.Wrap(err, "failed to get twin client"))
	}

	var msg rmb.Message
	err = json.Unmarshal(buffer.Bytes(), &msg)
	if err != nil {
		log.Error().Err(err).Msg("failed to load the message")
		return nil, mw.Error(errors.Wrap(err, "failed to load the message"))
	}

	err = p.resolver.Verify(msg.TwinSrc, &msg)
	if err != nil {
		log.Error().Err(err).Msg("failed to verify the message")
		return nil, mw.Error(errors.Wrap(err, "failed to verify the message"))
	}

	msgId, err := p.rmb.Submit(&msg)
	if err != nil {
		log.Error().Err(err).Msg("failed to submit the message")
		return nil, mw.BadGateway(errors.Wrap(err, "failed to submit the message"))
	}

	return msgId, nil
}

// getResult godoc
// @Summary Get the message result
// @Description Get the message result
// @Tags RMB
// @Accept  json
// @Produce  json
// @Param twin_id path int true "twin id"
// @Param retqueue path string true "message retqueue"
// @Success 200 {array} Message
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Failure 502 {object} string
// @Router /twin/{twin_id}/{retqueue} [get]
func (p *Proxy) getResult(r *http.Request) (interface{}, mw.Response) {
	twinIdString := mux.Vars(r)["twin_id"]
	ret := mux.Vars(r)["retqueue"]

	twinId, err := strconv.Atoi(twinIdString)
	if err != nil {
		return nil, mw.BadRequest(errors.Wrap(err, "invalid twin_id"))
	}

	_, err = p.resolver.Get(twinId)
	if err != nil {
		log.Error().Err(err).Msg("failed to get twin")
		return nil, mw.Error(errors.Wrap(err, "failed to get twin client"))
	}

	msgId := MessageIdentifier{
		Retqueue: ret,
	}

	messages, err := p.rmb.GetResult(msgId)
	if err != nil {
		log.Error().Err(err).Msg("couldn't get result")
		return nil, mw.BadGateway(errors.Wrap(err, "failed to get message"))
	}

	return messages, nil
}

// ping godoc
// @Summary ping the server
// @Description ping the server to check if it is running
// @Tags ping
// @Accept  json
// @Produce  json
// @Success 200 {object} PingMessage
// @Router /ping [get]
func (p *Proxy) ping(r *http.Request) (interface{}, mw.Response) {
	return map[string]string{"ping": "pong"}, mw.Ok()
}

// Setup : sets rmb routes
// @title RMB proxy API
// @version 1.0
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email soberkoder@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /
func (p *Proxy) Setup(router *mux.Router) {
	router.HandleFunc("/twin/{twin_id:[0-9]+}", mw.AsProxyHandlerFunc(p.sendMessage))
	router.HandleFunc("/twin/{twin_id:[0-9]+}/{retqueue}", mw.AsProxyHandlerFunc(p.getResult))
	router.HandleFunc("/ping", mw.AsHandlerFunc(p.ping))
	router.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)
}
