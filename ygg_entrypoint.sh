#!/usr/bin/env sh

set -ex

mkdir -p /dev/net
mknod /dev/net/tun c 10 200
chmod 600 /dev/net/tun

if [ ! -f "/etc/config.conf" ]; then
  echo "generate /etc/config.conf"
  yggdrasil --genconf > "/etc/config.conf"
fi

sed -i "/Peers: \[\]/c\  Peers: \n  [\n    tls:\/\/54.37.137.221:11129\n  ]" /etc/config.conf

yggdrasil --useconf < /etc/config.conf
exit $?
