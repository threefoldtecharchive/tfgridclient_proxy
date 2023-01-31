# The RMB

RMB is (reliable message bus) a set of tools (client and daemon) that aims to abstract inter-process communication between multiple processes running over multiple nodes.

For more about what is RMB, why for and how it is implemented you can check the repos for:

- the old deprecated rmb written in go: [rmb-go](https://github.com/threefoldtech/rmb_go).
- the new sill under-development rmb written in rust: [rmb-rs](https://github.com/threefoldtech/rmb-rs).

The RMB proxy (part of GridProxy project) is a client that can talk to remote RMB server. to run an RMB server check the repos above.
## How to run the RMB
run the RMB proxy and interact with the endpoints.
  ```bash
  make restart
  ```
  the server serve both explorer/rmb endpoints.

## RMB endpoints.

### V1

| HTTP Verb | Endpoint                   | Description                  |
| --------- | -------------------------- | ---------------------------- |
| POST      | `/twin/:twin_id`           | Send msg to the twin         |
| GET       | `/twin/:twin_id/:retqueue` | Get the result from the twin |

### V2

| HTTP Verb | Endpoint                          | Description                  |
| --------- | --------------------------------- | ---------------------------- |
| POST      | `/api/v2/twin/:twin_id`           | Send msg to the twin         |
| GET       | `/api/v2/twin/:twin_id/:retqueue` | Get the result from the twin |

### Message format

The message sent to the twin must contain some data about the destination, the process and extra info. it will look something like that:

```json
{
  "cmd": "zos.statistics.get",
  "dat": "",
  "dst": [2, 3],
  "err": "",
  "exp": 0,
  "now": 1631078674,
  "ret": "",
  "shm": "",
  "src": 1,
  "try": 2,
  "uid": "",
  "ver": 1
}
```

You can construct the message with the help of twin_server in Grid-Client check the [repo](https://github.com/threefoldtech/grid3_client_ts/blob/development/docs/rmb_server.md)
