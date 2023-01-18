package explorer

import (
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/threefoldtech/grid_proxy_server/internal/explorer/mw"
)

// getV2Contracts godoc
//
//	@Summary		Show contracts on the grid
//	@Description	Get all contracts on the grid, It has pagination
//	@Tags			GridProxy v2.0
//	@Accept			json
//	@Produce		json
//	@Param			page					query		int		false	"Page number"
//	@Param			size					query		int		false	"Max result per page"
//	@Param			ret_count				query		bool	false	"Set contracts' count on headers based on filter"
//	@Param			contract_id				query		int		false	"contract id"
//	@Param			twin_id					query		int		false	"twin id"
//	@Param			node_id					query		int		false	"node id which contract is deployed on in case of ('rent' or 'node' contracts)"
//	@Param			name					query		string	false	"contract name in case of 'name' contracts"
//	@Param			type					query		string	false	"contract type 'node', 'name', or 'rent'"
//	@Param			state					query		string	false	"contract state 'Created', 'GracePeriod', or 'Deleted'"
//	@Param			deployment_data			query		string	false	"contract deployment data in case of 'node' contracts"
//	@Param			deployment_hash			query		string	false	"contract deployment hash in case of 'node' contracts"
//	@Param			number_of_public_ips	query		int		false	"Min number of public ips in the 'node' contract"
//	@Success		200						{object}	[]types.Contract
//	@Failure		400						{object}	string
//	@Failure		500						{object}	string
//	@Router			/api/v2/contracts [get]
func (a *App) getV2Contracts(r *http.Request) (interface{}, mw.Response) {
	return a.listContracts(r)
}

// getV2Twins godoc
//
//	@Summary		Show twins on the grid
//	@Description	Get all twins on the grid, It has pagination
//	@Tags			GridProxy v2.0
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		false	"Page number"
//	@Param			size		query		int		false	"Max result per page"
//	@Param			ret_count	query		bool	false	"Set twins' count on headers based on filter"
//	@Param			twin_id		query		int		false	"twin id"
//	@Param			account_id	query		string	false	"account address"
//	@Success		200			{object}	[]types.Twin
//	@Failure		400			{object}	string
//	@Failure		500			{object}	string
//	@Router			/api/v2/twins [get]
func (a *App) getV2Twins(r *http.Request) (interface{}, mw.Response) {
	return a.listTwins(r)
}

// getV2Gateway godoc
//
//	@Summary		Show the details for specific gateway
//	@Description	Get all details for specific gateway hardware, capacity, DMI, hypervisor
//	@Tags			GridProxy v2.0
//	@Param			node_id	path	int	false	"Node ID"
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	types.Node
//	@Failure		400	{object}	string
//	@Failure		404	{object}	string
//	@Failure		500	{object}	string
//	@Router			/api/v2/gateways/{node_id} [get]
func (a *App) getV2Gateway(r *http.Request) (interface{}, mw.Response) {
	return a.loadNode(r)
}

// getV2Node godoc
//
//	@Summary		Show the details for specific node
//	@Description	Get all details for specific node hardware, capacity, DMI, hypervisor
//	@Tags			GridProxy v2.0
//	@Param			node_id	path	int	false	"Node ID"
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	types.Node
//	@Failure		400	{object}	string
//	@Failure		404	{object}	string
//	@Failure		500	{object}	string
//	@Router			/api/v2/nodes/{node_id} [get]
func (a *App) getV2Node(r *http.Request) (interface{}, mw.Response) {
	return a.loadNode(r)
}

// getV2Gateways godoc
//
//	@Summary		Show gateways on the grid
//	@Description	Get all gateways on the grid, It has pagination. Nodes displayed with nested capacity object
//	@Tags			GridProxy v2.0
//	@Accept			json
//	@Produce		json
//	@Param			page			query		int		false	"Page number"
//	@Param			size			query		int		false	"Max result per page"
//	@Param			ret_count		query		bool	false	"Set nodes' count on headers based on filter"
//	@Param			free_mru		query		int		false	"Min free reservable mru in bytes"
//	@Param			free_hru		query		int		false	"Min free reservable hru in bytes"
//	@Param			free_sru		query		int		false	"Min free reservable sru in bytes"
//	@Param			free_ips		query		int		false	"Min number of free ips in the farm of the node"
//	@Param			status			query		string	false	"Node status filter, 'up': for only up nodes & 'down': for all up/down nodes."
//	@Param			city			query		string	false	"Node city filter"
//	@Param			country			query		string	false	"Node country filter"
//	@Param			farm_name		query		string	false	"Get nodes for specific farm"
//	@Param			ipv4			query		bool	false	"Set to true to filter nodes with ipv4"
//	@Param			ipv6			query		bool	false	"Set to true to filter nodes with ipv6"
//	@Param			domain			query		bool	false	"Set to true to filter nodes with domain"
//	@Param			dedicated		query		bool	false	"Set to true to get the dedicated nodes only"
//	@Param			rentable		query		bool	false	"Set to true to filter the available nodes for renting"
//	@Param			rented			query		bool	false	"Set to true to filter rented nodes"
//	@Param			rented_by		query		int		false	"rented by twin id"
//	@Param			available_for	query		int		false	"available for twin id"
//	@Param			farm_ids		query		string	false	"List of farms separated by comma to fetch nodes from (e.g. '1,2,3')"
//	@Success		200				{object}	[]types.NodeWithNestedCapacity
//	@Failure		400				{object}	string
//	@Failure		500				{object}	string
//	@Router			/api/v2/gateways [get]
func (a *App) getV2Gateways(r *http.Request) (interface{}, mw.Response) {
	return a.listNodesWithNestedCapacity(r)
}

// getV2Nodes godoc
//
//	@Summary		Show nodes on the grid
//	@Description	Get all nodes on the grid, It has pagination. Nodes displayed with nested capacity object
//	@Tags			GridProxy v2.0
//	@Accept			json
//	@Produce		json
//	@Param			page			query		int		false	"Page number"
//	@Param			size			query		int		false	"Max result per page"
//	@Param			ret_count		query		bool	false	"Set nodes' count on headers based on filter"
//	@Param			free_mru		query		int		false	"Min free reservable mru in bytes"
//	@Param			free_hru		query		int		false	"Min free reservable hru in bytes"
//	@Param			free_sru		query		int		false	"Min free reservable sru in bytes"
//	@Param			free_ips		query		int		false	"Min number of free ips in the farm of the node"
//	@Param			status			query		string	false	"Node status filter, 'up': for only up nodes & 'down': for all up/down nodes."
//	@Param			city			query		string	false	"Node city filter"
//	@Param			country			query		string	false	"Node country filter"
//	@Param			farm_name		query		string	false	"Get nodes for specific farm"
//	@Param			ipv4			query		bool	false	"Set to true to filter nodes with ipv4"
//	@Param			ipv6			query		bool	false	"Set to true to filter nodes with ipv6"
//	@Param			domain			query		bool	false	"Set to true to filter nodes with domain"
//	@Param			dedicated		query		bool	false	"Set to true to get the dedicated nodes only"
//	@Param			rentable		query		bool	false	"Set to true to filter the available nodes for renting"
//	@Param			rented			query		bool	false	"Set to true to filter rented nodes"
//	@Param			rented_by		query		int		false	"rented by twin id"
//	@Param			available_for	query		int		false	"available for twin id"
//	@Param			farm_ids		query		string	false	"List of farms separated by comma to fetch nodes from (e.g. '1,2,3')"
//	@Success		200				{object}	[]types.NodeWithNestedCapacity
//	@Failure		400				{object}	string
//	@Failure		500				{object}	string
//	@Router			/api/v2/nodes [get]
func (a *App) getV2Nodes(r *http.Request) (interface{}, mw.Response) {
	return a.listNodesWithNestedCapacity(r)
}

// getV2Stats godoc
//
//	@Summary		Show stats about the grid
//	@Description	Get statistics about the grid
//	@Tags			GridProxy v2.0
//	@Accept			json
//	@Produce		json
//	@Param			status	query		string	false	"Node status filter, 'up': for only up nodes & 'down': for all up/down nodes."
//	@Success		200		{object}	[]types.Counters
//	@Failure		400		{object}	string
//	@Failure		500		{object}	string
//	@Router			/api/v2/stats [get]
func (a *App) getV2Stats(r *http.Request) (interface{}, mw.Response) {
	return a.loadStats(r)
}

// getV2Farms godoc
//
//	@Summary		Show farms on the grid
//	@Description	Get all farms on the grid, It has pagination
//	@Tags			GridProxy v2.0
//	@Accept			json
//	@Produce		json
//	@Param			page				query		int		false	"Page number"
//	@Param			size				query		int		false	"Max result per page"
//	@Param			ret_count			query		bool	false	"Set farms' count on headers based on filter"
//	@Param			free_ips			query		int		false	"Min number of free ips in the farm"
//	@Param			total_ips			query		int		false	"Min number of total ips in the farm"
//	@Param			pricing_policy_id	query		int		false	"Pricing policy id"
//	@Param			version				query		int		false	"farm version"
//	@Param			farm_id				query		int		false	"farm id"
//	@Param			twin_id				query		int		false	"twin id associated with the farm"
//	@Param			name				query		string	false	"farm name"
//	@Param			name_contains		query		string	false	"farm name contains"
//	@Param			certification_type	query		string	false	"certificate type DIY or Certified"
//	@Param			dedicated			query		bool	false	"farm is dedicated"
//	@Param			stellar_address		query		string	false	"farm stellar_address"
//	@Success		200					{object}	[]types.Farm
//	@Failure		400					{object}	string
//	@Failure		500					{object}	string
//	@Router			/api/v2/farms [get]
func (a *App) getV2Farms(r *http.Request) (interface{}, mw.Response) {
	return a.listFarms(r)
}

// Setup is the server and do initial configurations
//
//	@title			Grid Proxy Server API
//	@version		1.0
//	@description	grid proxy server has the main methods to list farms, nodes, node details in the grid.
//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html
//	@BasePath		/api/v2
func (a *App) loadV2Handlers(router *mux.Router) {
	router.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)

	router.HandleFunc("/", mw.AsHandlerFunc(a.indexPage(router)))
	router.HandleFunc("/version", mw.AsHandlerFunc(a.version))
	router.HandleFunc("/stats", mw.AsHandlerFunc(a.getV2Stats))

	router.HandleFunc("/farms", mw.AsHandlerFunc(a.getV2Farms))
	router.HandleFunc("/nodes", mw.AsHandlerFunc(a.getV2Nodes))
	router.HandleFunc("/gateways", mw.AsHandlerFunc(a.getV2Gateways))
	router.HandleFunc("/twins", mw.AsHandlerFunc(a.getV2Twins))
	router.HandleFunc("/contracts", mw.AsHandlerFunc(a.getV2Contracts))
	router.HandleFunc("/nodes/{node_id:[0-9]+}", mw.AsHandlerFunc(a.getV2Node))
	router.HandleFunc("/gateways/{node_id:[0-9]+}", mw.AsHandlerFunc(a.getV2Gateway))
	router.HandleFunc("/nodes/{node_id:[0-9]+}/status", mw.AsHandlerFunc(a.getNodeStatus))
	router.HandleFunc("/gateways/{node_id:[0-9]+}/status", mw.AsHandlerFunc(a.getNodeStatus))
}
