#!/usr/bin/env sh

set -ex

if [ ! -f "/yggdrasil/yggdrasil.conf" ]; then
  echo "generating new configurations at /yggdrasil/yggdrasil.conf"
  mkdir -p /yggdrasil/
  yggdrasil --genconf > "/yggdrasil/yggdrasil.conf"
  sed -i "/Peers: \[\]/c\  Peers: \n  [\n    tls:\/\/54.37.137.221:11129\n  ]" /yggdrasil/yggdrasil.conf
fi

yggdrasil --useconf < /yggdrasil/yggdrasil.conf
exit $?
