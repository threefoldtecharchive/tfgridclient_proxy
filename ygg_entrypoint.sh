#!/usr/bin/env sh

set -ex

CONF_DIR="/etc/yggdrasil-network"

if [ ! -f "$CONF_DIR/config.conf" ]; then
  echo "generate $CONF_DIR/config.conf"
  yggdrasil --genconf > "$CONF_DIR/config.conf"
fi

sed -i "/Peers: \[\]/c\  Peers: \n  [\n    tls:\/\/54.37.137.221:11129\n  ]" $CONF_DIR/config.conf

yggdrasil --useconf < $CONF_DIR/config.conf
exit $?