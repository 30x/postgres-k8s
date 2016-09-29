#!/bin/bash


function get_host_at_index(){
  #Split the FQDN into parts based on the '.' char
  IFS='.' read -ra hostcomponents <<< "$FQDN"

  #Get all indexes, so we can iterate them and know the index
  indexes=${!hostcomponents[*]}

  # echo "Indexes is ${indexes}"

  output=""
  indexOut=$1

  for index in $indexes;
  do
    part=${hostcomponents[$index]}

    # echo "part is ${part}"

    #First index, extract the hostname
    if [ $index -eq 0 ]; then
      output=${part%-*}
      output="${output}-${indexOut}"
    else
      output="${output}.$part"
    fi



  done

  echo $output


}

#Configure the master
function configure_master() {
  echo "Configuring master"

  echo "Configuring the archive directory"

  mkdir -p $PGDATA/archive
  chown -R postgres $PGDATA



  allowed_replicas="10.244.0.0/16"
  slave_node=$(get_host_at_index 1)

  echo "Configuring $PGDATA/pg_hba.conf for master"
  cat << EOF >> $PGDATA/pg_hba.conf
host	replication	postgres	${allowed_replicas}	trust
EOF


  echo "Configuring $PGDATA/postgresql.conf for master"
  cat << EOF >> $PGDATA/postgresql.conf
#-------------------------
# START K8S AUTO CONFIGURE
#-------------------------
wal_level = hot_standby
archive_mode = on
archive_command = 'test ! -f $PGDATA/archive/%f && cp %p $PGDATA/archive/%f'
max_wal_senders = 3
max_replication_slots = 3
#synchronous_standby_names = '${slave_node}'
synchronous_standby_names = '10.244.0.7'
#-------------------------
# END K8S AUTO CONFIGURE
#-------------------------
EOF


#   #Rendered from pod environment
# host	replication	postgres	%s	trust
# `
  # cat ${pg_conf} >> $PGDATA/pg_hba.conf
}



function configure_slave() {
  echo "Configuring slave"

  #Copy the original pg configuration over
  # cp /var/lib/postgres/data/postgresql.conf /tmp/postgresql.conf

  echo "Stopping postgres to being the backup"


  #stop postgres before configuring
  gosu postgres pg_ctl stop -D $PGDATA/

  #Data is a mount point, so remove everything under it
  rm -rf $PGDATA/*
  #Make sure ownership is correct
  chown -R postgres $PGDATA

  #configure the system

  master_node=$(get_host_at_index 0)

  echo "Beginning bootstrap from host ${master_node}"

  gosu postgres pg_basebackup -D $PGDATA -p 5432 -U postgres -v -h ${master_node} --xlog-method=stream

  if [ $? -ne 0 ]; then
    echo "FAILURE: Unable to restore from backup, existing"
    exit 1
  fi



  echo "Configuring $PGDATA/postgresql.conf for slave"


  #DELETE the configuration from the master if we're doing sync replication
  sed -i -- '/.*START K8S/,/.*END K8S/d'  $PGDATA/postgresql.conf

  cat << EOF >> $PGDATA/postgresql.conf
hot_standby = on
EOF

  echo "Configuring recovery"
  cp /usr/share/postgresql/9.5/recovery.conf.sample $PGDATA/recovery.conf

  cat << EOF >> $PGDATA/recovery.conf
  standby_mode = on
  primary_conninfo = 'host=${master_node} port=5432 user=postgres'
EOF


  #Make sure postgres owns everythign, or it will crap out
  chown -R postgres  /var/lib/postgresql


  #copy over the original postgres conf
  # cp /tmp/postgresql.conf /var/lib/postgres/data/postgresql.conf

  # echo ""

  #Stop postgres to avoid postmaster.pid not found errors
  # su -c "pg_ctl stop -D $PGDATA/" postgres
  #
  #Start postgres
    echo "Starting postgres"
    gosu postgres pg_ctl -D "$PGDATA" \
			-o "-c listen_addresses='localhost'" \
			-w start


    echo "Done configuring slave"
}

function configure_replica() {
  echo "Configuring replica"
  su -c "pg_ctl stop -D $PGDATA/" postgres

  cp /var/lib/postgres/data/postgresql.conf /tmp/postgresql.conf

  slave_node=$(get_host_at_index 1)

  su -c "pg_basebackup -D $PGDATA -p 5432 -U postgres -v -h ${slave_node} --xlog-method=stream" postgres


    if [ $? -ne 0 ]; then
      echo "FAILURE: Unable to restore from backup, existing"
      exit 1
    fi


    #Make sure postgres owns everythign, or it will crap out
    chown -R postgres  /var/lib/postgresql

  # /postgres-config/posgres-agent replica configure --data $PGDATA/  --hostname $FQDN --port 5432 --user postgres --pg_conf $PGDATA/postgresql.conf
  su -c "pg_ctl start -D $PGDATA/" postgres
}


#get the short hostname for comparison
FQDN=$(hostname -f)
#
# HOSTNAME="foo-1"
# FQDN="foo-1.bar.baz.local"


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
