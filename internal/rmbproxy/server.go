package rmbproxy

import (
	"bytes"
	"net/http"
	"strconv"

	// swagger configuration
	"github.com/pkg/errors"
	"github.com/threefoldtech/grid_proxy_server/internal/explorer/mw"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

func (a *App) sendMessage(r *http.Request) (*http.Response, mw.Response) {
	twinIDString := mux.Vars(r)["twin_id"]

	buffer := new(bytes.Buffer)
	if _, err := buffer.ReadFrom(r.Body); err != nil {
		return nil, mw.BadRequest(err)
	}

	twinID, err := strconv.Atoi(twinIDString)
	if err != nil {
		return nil, mw.BadRequest(errors.Wrap(err, "invalid twin_id"))
	}

	c, err := a.resolver.Get(twinID)
	if err != nil {
		log.Error().Err(err).Msg("failed to create twin client")
		return nil, mw.Error(errors.Wrap(err, "failed to create twin client"))
	}

	response, err := c.SubmitMessage(*buffer)
	if err != nil {
		return nil, mw.BadGateway(errors.Wrap(err, "failed to submit message"))
	}
	return response, nil
}

func (a *App) getResult(r *http.Request) (*http.Response, mw.Response) {
	twinIDString := mux.Vars(r)["twin_id"]
	retqueue := mux.Vars(r)["retqueue"]

	reqBody := MessageIdentifier{
		Retqueue: retqueue,
	}

	twinID, err := strconv.Atoi(twinIDString)
	if err != nil {
		return nil, mw.BadRequest(errors.Wrap(err, "invalid twin_id"))
	}

	c, err := a.resolver.Get(twinID)
	if err != nil {
		log.Error().Err(err).Msg("failed to create twin client")
		return nil, mw.Error(errors.Wrap(err, "failed to create twin client"))
	}

	response, err := c.GetResult(reqBody)
	if err != nil {
		return nil, mw.BadGateway(errors.Wrap(err, "failed to submit message"))
	}
	return response, nil
}

func (a *App) ping(r *http.Request) (interface{}, mw.Response) {
	return PingMessage{Ping: "pong"}, mw.Ok()
}

func Setup(version string, router *mux.Router, resolver *TwinExplorerResolver) error {
	a := &App{
		resolver: resolver,
	}

	if version == "v1" {
		a.loadV1Handlers(router)
	} else {
		a.loadV2Handlers(router)
	}

	return nil
}
