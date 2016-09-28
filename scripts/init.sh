#!/bin/bash


function get_host_at_index(){
  #Split the FQDN into parts based on the '.' char
  IFS='.' read -ra hostcomponents <<< "$FQDN"

  #Get all indexes, so we can iterate them and know the index
  indexes=${!hostcomponents[*]}

  echo "Indexes is ${indexes}"
  
  output=""
  indexOut=$1

  for index in $indexes;
  do
    part=${indexes[$index]}

    echo ${part}

    #First index, extract the hostname
    if [ $index -eq 0 ]; then
      output=${part%-*}
      output="${output}-${indexOut}"
    fi

    output="${output}.part"

  done

  echo $output


}

#Configure the master
function configure_master() {
  echo "Configuring master"

  ARCHIVE_DIR="/var/lib/postgresql"
  mkdir -p /var/lib/postgresql/data/archive
  chown -R postgres $ARCHIVE_DIR


  allowed_replicas="10.244.0.0/16"
  slave_name=$(get_host_at_index 1)

  cat << EOF >> /var/lib/postgresql/data/pg_hba.conf
  host	replication	postgres	${allowed_replicas}	trust
EOF


  cat << EOF >> /var/lib/postgresql/data/postgresql.conf
  wal_level = hot_standby
  archive_mode = on
  archive_command = 'test ! -f /var/lib/postgresql/data/archive/%f && cp %p /var/lib/postgresql/data/archive/%f'
  max_wal_senders = 3
  synchronous_standby_names = '${slave_name}'
EOF


#   #Rendered from pod environment
# host	replication	postgres	%s	trust
# `
  # cat ${pg_conf} >> /var/lib/postgresql/data/pg_hba.conf
}



function configure_slave() {
   echo "Configuring slave"

  #Copy the original pg configuration over
  cp /var/lib/postgres/data/postgresql.conf /tmp/postgresql.conf

  #stop postgres before configuring
  su -c "/usr/lib/postgresql/9.5/bin/pg_ctl stop -D /var/lib/postgresql/data/" postgres

  #configure the system

  master_node=$(/postgres-config/posgres-agent slave configure --hostname $FQDN)

  su -c "/usr/lib/postgresql/9.5/bin/pg_basebackup -D /var/lib/postgresql/data -p 5432 -U postgres -v -h ${master_node} --xlog-method=stream" postgres

  if [ $? -ne 0 ]; then
    echo "FAILURE: Unable to restore from backup, existing"
    exit 1
  fi


  #Make sure postgres owns everythign, or it will crap out
  chown -R postgres  /var/lib/postgresql

  #copy over the original postgres conf
  # cp /tmp/postgresql.conf /var/lib/postgres/data/postgresql.conf

  # echo ""

  #Start postgres
  su -c "/usr/lib/postgresql/9.5/bin/pg_ctl start -D /var/lib/postgresql/data/" postgres
}

function configure_replica() {
  echo "Configuring replica"
  su -c "/usr/lib/postgresql/9.5/bin/pg_ctl stop -D /var/lib/postgresql/data/" postgres
  # /postgres-config/posgres-agent replica configure --data /var/lib/postgresql/data/  --hostname $FQDN --port 5432 --user postgres --pg_conf /var/lib/postgresql/data/postgresql.conf
  su -c "/usr/lib/postgresql/9.5/bin/pg_ctl start -D /var/lib/postgresql/data/" postgres
}


#get the short hostname for comparison
FQDN=$(hostname -f)

HOSTNAME="foo-0"
FQDN="foo-0.local"


echo "Configuring postres"
echo "Hostname is $HOSTNAME"
echo "FQDN is $FQDN"

#Make sure the postgres process owns the mounted directory


if [[ "$HOSTNAME" == *-0 ]]; then
  configure_master
elif [[ "$HOSTNAME" == *-1  ]]; then
  configure_slave
else
  configure_replica
fi
