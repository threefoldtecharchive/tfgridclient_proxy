package rmbproxy

import (
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/threefoldtech/grid_proxy_server/internal/explorer/mw"
)

// sendV2 godoc
//
//	@Summary		submit the message
//	@Description	submit the message
//	@Tags			RMB v2.0
//	@Accept			json
//	@Produce		json
//	@Param			msg		body		Message	true	"rmb.Message"
//	@Param			twin_id	path		int		true	"twin id"
//	@Success		200		{object}	MessageIdentifier
//	@Failure		400		{object}	string
//	@Failure		500		{object}	string
//	@Failure		502		{object}	string
//	@Router			/api/v2/twin/{twin_id} [post]
func (a *App) sendV2(r *http.Request) (*http.Response, mw.Response) {
	return a.sendMessage(r)
}

// getV2 godoc
//
//	@Summary		Get the message result
//	@Description	Get the message result
//	@Tags			RMB v2.0
//	@Accept			json
//	@Produce		json
//	@Param			twin_id		path		int		true	"twin id"
//	@Param			retqueue	path		string	true	"message retqueue"
//	@Success		200			{array}		Message
//	@Failure		400			{object}	string
//	@Failure		500			{object}	string
//	@Failure		502			{object}	string
//	@Router			/api/v2/twin/{twin_id}/{retqueue} [get]
func (a *App) getV2(r *http.Request) (*http.Response, mw.Response) {
	return a.getResult(r)
}

// pingServerV2 godoc
//
//	@Summary		ping the server
//	@Description	ping the server to check if it is running
//	@Tags			Ping v2.0
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	PingMessage
//	@Router			/api/v2/ping [get]
func (a *App) pingServerV2(r *http.Request) (interface{}, mw.Response) {
	return a.getResult(r)
}

// Setup : sets rmb routes
//
//	@title			RMB proxy API
//	@version		2.0
//	@termsOfService	http://swagger.io/terms/
//	@contact.name	API Support
//	@contact.email	soberkoder@swagger.io
//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html
//	@host			localhost:8080
//	@BasePath		/api/v2
func (a *App) loadV2Handlers(router *mux.Router) {
	router.HandleFunc("/twin/{twin_id:[0-9]+}", mw.AsProxyHandlerFunc(a.sendV2))
	router.HandleFunc("/twin/{twin_id:[0-9]+}/{retqueue}", mw.AsProxyHandlerFunc(a.getV2))
	router.HandleFunc("/ping", mw.AsHandlerFunc(a.pingServerV2))
	router.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)
}
