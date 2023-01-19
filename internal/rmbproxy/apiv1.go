package rmbproxy

import (
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/threefoldtech/grid_proxy_server/internal/explorer/mw"
)

type ApiV1 struct {
	*App
}

// sendMessage godoc
//
//	@Summary		submit the message
//	@Description	submit the message
//	@Tags			RMB v1.0
//	@Accept			json
//	@Produce		json
//	@Param			msg		body		Message	true	"rmb.Message"
//	@Param			twin_id	path		int		true	"twin id"
//	@Success		200		{object}	MessageIdentifier
//	@Failure		400		{object}	string
//	@Failure		500		{object}	string
//	@Failure		502		{object}	string
//	@Router			/twin/{twin_id} [post]
func (a *ApiV1) sendMessage(r *http.Request) (*http.Response, mw.Response) {
	return a.SendMessage(r)
}

// getResult godoc
//
//	@Summary		Get the message result
//	@Description	Get the message result
//	@Tags			RMB v1.0
//	@Accept			json
//	@Produce		json
//	@Param			twin_id		path		int		true	"twin id"
//	@Param			retqueue	path		string	true	"message retqueue"
//	@Success		200			{array}		Message
//	@Failure		400			{object}	string
//	@Failure		500			{object}	string
//	@Failure		502			{object}	string
//	@Router			/twin/{twin_id}/{retqueue} [get]
func (a *ApiV1) getResult(r *http.Request) (*http.Response, mw.Response) {
	return a.GetResult(r)
}

// ping godoc
//
//	@Summary		ping the server
//	@Description	ping the server to check if it is running
//	@Tags			Ping v1.0
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	PingMessage
//	@Router			/ping [get]
func (a *ApiV1) ping(r *http.Request) (interface{}, mw.Response) {
	return a.Ping(r)
}

// Setup : sets rmb routes
//
//	@title			RMB proxy API
//	@version		1.0
//	@termsOfService	http://swagger.io/terms/
//	@contact.name	API Support
//	@contact.email	soberkoder@swagger.io
//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html
//	@host			localhost:8080
//	@BasePath		/
func (a *App) loadV1Handlers(router *mux.Router) {
	api := ApiV1{App: a}
	router.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)
	router.HandleFunc("/ping", mw.AsHandlerFunc(api.ping))

	router.HandleFunc("/twin/{twin_id:[0-9]+}", mw.AsProxyHandlerFunc(api.sendMessage))
	router.HandleFunc("/twin/{twin_id:[0-9]+}/{retqueue}", mw.AsProxyHandlerFunc(api.getResult))
}
