#!/bin/bash

SCRIPT_PATH="/nodesetup"

#get the short hostname for comparison
FQDN=$(hostname -f)



echo "Configuring postres"
echo "Hostname is $HOSTNAME"
echo "FQDN is $FQDN"

#Make sure the postgres process owns the mounted directory
mkdir -p /var/lib/postgresql/data/archive
chown -R postgres  /var/lib/postgresql

if [[ "$HOSTNAME" == *-0 ]]; then
  echo "Configuring master"
  
  /postgres-config/posgres-agent master configure --data /var/lib/postgresql/data/ --pg_conf /var/lib/postgresql/data/postgresql.conf --pg_hba /var/lib/postgresql/data/pg_hba.conf --hostname $FQDN
elif [[ "$HOSTNAME" == *-1  ]]; then
  echo "Configuring slave"

  #stop postgres before configuring
  su -c "/usr/lib/postgresql/9.5/bin/pg_ctl stop -D /var/lib/postgresql/data/" postgres
  
  #configure the system 
  /postgres-config/posgres-agent slave configure --data /var/lib/postgresql/data/  --hostname $FQDN --port 5432 --user postgres --pg_conf /var/lib/postgresql/data/postgresql.conf
  
  #Start postgres
  su -c "/usr/lib/postgresql/9.5/bin/pg_ctl start -D /var/lib/postgresql/data/" postgres

else
  echo "Configuring replica"
  su -c "/usr/lib/postgresql/9.5/bin/pg_ctl stop -D /var/lib/postgresql/data/" postgres
  /postgres-config/posgres-agent replica configure --data /var/lib/postgresql/data/  --hostname $FQDN --port 5432 --user postgres --pg_conf /var/lib/postgresql/data/postgresql.conf
  su -c "/usr/lib/postgresql/9.5/bin/pg_ctl start -D /var/lib/postgresql/data/" postgres
fi
