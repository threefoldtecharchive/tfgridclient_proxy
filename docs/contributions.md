# Contributions guide

## Project structure

The main structure of the code base is as follows:

- `charts`: helm chart
- `cmds`: includes the project Golang entrypoints
- `docs`: project documentation
- `internal`: contains the explorer API logic and the cert manager implementation, this where most of the feature work will be done
- `pkg`: contains client implementation and shared libs
- `tests`: integration tests
- `tools`: DB tools to prepare the Postgres DB for testing and development
- `rootfs`: ZOS root endpoint that will be mounted in the docker image

### Internal

- `explorer`: contains the explorer server logic:
  - `db`: the db connection and operations
  - `mw`: defines the generic action mount that will be be used as http handler
- `certmanager`: logic to ensure certificates are available and up to date

`server.go` includes the logic for all the API operations.

### Pkg

- `client`: client implementation
- `types`: defines all the API objects

## Writing tests

Adding a new endpoint should be accompanied with a corresponding test. Ideally every change or bug fix should include a test to ensure the new behavior/fix is working as intended.

Since these are integration tests, you need to first make sure that your local db is already seeded with the ncessary data. See tools [doc](../tools/db/README.md) for more information about how to prepare your db.

Testing tools offer two clients that are the basic of most tests:

- `local`: this client connects to the local db
- `proxy client`: this client connects to the running local instance

You need to start an instance of the server before running the tests. Check [here](commands.md) for how to start.
