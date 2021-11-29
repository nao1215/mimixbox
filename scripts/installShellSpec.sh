#!/bin/bash -eu

curl -fsSL https://git.io/shellspec > install.sh
chmod a+x install.sh

expect -c "
  set timeout 3
  spawn ./install.sh
  expect \"\[y\/N\]\"
  send \"y\n\"
  interact
"