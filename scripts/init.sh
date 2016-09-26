#!/bin/bash

SCRIPT_PATH="/nodesetup"

#get the short hostname for comparison
HOSTNAME=$(hostname -s)

echo "Configuring postres"
echo "Hostname is $HOSTNAME"

if [[ "$HOSTNAME" == *-0 ]]; then
  echo "Configuring master"
  /postgres-config/posgres-agent master configure --data /var/lib/postgresql/data/ --pg_conf /var/lib/postgresql/data/postgresql.conf --pg_hba /var/lib/postgresql/data/pg_hba.conf --hostname $(hostname)
elif [[ "$HOSTNAME" == *-1  ]]; then
  echo "Configuring slave"
  /postgres-config/posgres-agent slave configure --data /var/lib/postgresql/data/  --hostname $(hostname) --port 5432 --user postgres
else
  echo "Configuring replica"
fi