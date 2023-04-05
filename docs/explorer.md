# The Grid Explorer

A REST API used to index a various information from the TF Chain.

## How the explorer work?

- Due to limitations on indexing information from the blockchain, Complex inter-tables queries and limitations can't be applied directly on the chain.
- Here comes the TFGridDB, a shadow database contains all the data on the chain which is being updated each 2 hours.
- Then the explorer can apply a raw SQL queries on the database with all limitations and filtration needed.
- The used technology to extract the info from the blockchain is Subsquid check the [repo](https://github.com/threefoldtech/tfchain_graphql)

## Explorer Endpoints

| HTTP Verb | Endpoint                    | Description                        |
| --------- | --------------------------- | ---------------------------------- |
| GET       | `/contracts`                | Show all contracts on the chain    |
| GET       | `/farms`                    | Show all farms on the chain        |
| GET       | `/gateways`                 | Show all gateway nodes on the grid |
| GET       | `/gateways/:node_id`        | Get a single gateway node details  |
| GET       | `/gateways/:node_id/status` | Get a single node status           |
| GET       | `/nodes`                    | Show all nodes on the grid         |
| GET       | `/nodes/:node_id`           | Get a single node details          |
| GET       | `/nodes/:node_id/status`    | Get a single node status           |
| GET       | `/stats`                    | Show the grid statistics           |
| GET       | `/twins`                    | Show all the twins on the chain    |
| GET       | `/nodes/:node_id/statistics`| Get a single node ZOS statistics   |

For the available filters on each node. check `/swagger/index.html` endpoint on the running instance.
