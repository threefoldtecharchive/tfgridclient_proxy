# Setup Prerequisites

## Golang

Install Golang official [guide](https://go.dev/doc/install).

## Docker

Docker is useful for running the TFGridDb in container environment. Install Docker engine Digital ocean [guide](https://www.digitalocean.com/community/tutorials/how-to-install-and-use-docker-on-ubuntu-22-04).

Note: it will be necessary to follow step #2 in the previous article to run docker without sudo. if you want to avoid that. edit the docker commands in the `Makefile` and add sudo.

## Postgres

If you have docker installed you can run postgres on a container with:

```bash
make db-start
```

Then you can either load a dump of the database if you have one:

```bash
make db-dump p=~/dump.sql
```

or easier you can fill the database tables with randomly generated data with the script `tools/db/generate.go` to do that run:

```bash
make db-fill
```

## Get Mnemonics

1. Install [polkadot extension](https://github.com/polkadot-js/extension) on your browser.
2. Create a new account from the extension. It is important to save the seeds.
