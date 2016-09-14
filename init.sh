#!/bin/bash

SCRIPT_PATH="/nodesetup"

#get the short hostname for comparison
HOSTNAME=$(hostname -s)

echo "Hostname is $HOSTNAME"

if [[ "$HOSTNAME" == *-0 ]]; then
    echo "Master Node"
elif [[ "$HOSTNAME" == *-1  ]]; then
  echo "Slave Node"
else
  echo "Replica Node"
fi
