#!/usr/bin/env sh

set -ex

if [ ! -f "/etc/yggdrasil.conf" ]; then
  echo "generating new configurations at /etc/yggdrasil.conf"
  yggdrasil --genconf > "/etc/yggdrasil.conf"
  sed -i "/Peers: \[\]/c\  Peers: \n  [\n    tls:\/\/54.37.137.221:11129\n $PEERS  ]" /etc/yggdrasil.conf
  sed -i "/^  PublicKey: */c\  PublicKey: $PUBLIC_KEY" /etc/yggdrasil.conf
  sed -i "/PrivateKey: */c\  PrivateKey: $PRIVATE_KEY" /etc/yggdrasil.conf
fi

yggdrasil --useconf < /etc/yggdrasil.conf
exit $?
