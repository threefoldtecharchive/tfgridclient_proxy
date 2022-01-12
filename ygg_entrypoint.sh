#!/usr/bin/env sh

set -ex


if [ ! -f "/yggdrasil/yggdrasil.conf" ]; then
  echo "generate /yggdrasil/yggdrasil.conf"
  yggdrasil --genconf > "/yggdrasil/yggdrasil.conf"
fi

sed -i "/Peers: \[\]/c\  Peers: \n  [\n    tls:\/\/54.37.137.221:11129\n  ]" /yggdrasil/yggdrasil.conf

yggdrasil --useconf < /yggdrasil/yggdrasil.conf
exit $?
