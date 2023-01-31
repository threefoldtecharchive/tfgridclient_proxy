# The Grid Explorer

A REST API used to index a various information from the TF Chain.

## How the explorer work?

- Due to limitations on indexing information from the blockchain, Complex inter-tables queries and limitations can't be applied directly on the chain.
- Here comes the TFGridDB, a shadow database contains all the data on the chain which is being updated each 2 hours.
- Then the explorer can apply a raw SQL queries on the database with all limitations and filtration needed.
- The used technology to extract the info from the blockchain is Subsquid check the [repo](https://github.com/threefoldtech/tfchain_graphql)

## Explorer Endpoints

### V1

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

For the available filters on each node. check `/swagger/index.html` endpoint on the running instance.

### V2

| HTTP Verb | Endpoint                        | Description                     |
| --------- | ------------------------------- | ------------------------------- |
| GET       | `/api/v2/contracts`             | Show all contracts on the chain |
| GET       | `/api/v2/farms`                 | Show all farms on the chain     |
| GET       | `/api/v2/nodes`                 | Show all nodes on the grid      |
| GET       | `/api/v2/nodes/:node_id`        | Get a single node details       |
| GET       | `/api/v2/nodes/:node_id/status` | Get a single node status        |
| GET       | `/api/v2/stats`                 | Show the grid statistics        |
| GET       | `/api/v2/twins`                 | Show all the twins on the chain |

For the changes introduced on the `/api/v2` check the [CHANGELOG](./docs/CHANGELOG.md).


For the available filters on each node. check `/api/v2/swagger/index.html` endpoint on the running instance.
