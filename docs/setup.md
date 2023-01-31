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

## Get Yggdrasil IP

Check the official [Guide](https://yggdrasil-network.github.io/installation-linux-deb.html).
Or Threefold [manual](https://library.threefold.me/info/manual/#/manual__yggdrasil_client?id=new-peer-list-for-usage-in-every-yggdrasil-planetary-network-client)

1. Install
   ```bash
   sudo apt-get install yggdrasil
   ```
2. Generate the conf file if not there
   after installation a `/etc/yggdrasil.conf` file should be created, in case it didn't
   ```bash
   yggdrasil -genconf > /etc/yggdrasil.conf
   ```
3. Configure the peers
   Add these [peers](https://github.com/threefoldtech/zos-config/blob/main/production.json) which is tracked by the operation team to the peers list in the conf file for a better communication with the other grid nodes.
4. Reload the service & enable to run at startup
   ```bash
   sudo systemctl enable yggdrasil
   sudo systemctl start yggdrasil
   ```

- Common problem: `systemctl start yggdrasil` maybe failed because of a non-existing config file. if that the case edit the ExecCommand in `/lib/systemd/system/yggdrasil.service` to use the correct path for the config file. then
    ```bash
    systemctl daemon-reload
    ```

## Get Mnemonics
1. Install [polkadot extension](https://github.com/polkadot-js/extension) on your browser.
2. Create a new account from the extension. It is important to save the seeds.

## Get Chain Twin
After you get the mnemonics and successfully run yggdrasil on your machine.
- On a terminal run 
    ```bash
    sudo yggdrasilctl getSelf
    ```
    and copy your `IPv6 address`
- Go to [Dashboard](https://dashboard.dev.grid.tf/). and sign in with your account.
- Edit your twin details and add your `Ygg IP`. now you have an identity on the chain.

