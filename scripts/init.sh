#!/bin/bash



#Configure the master
function configure_master() {
  echo "Configuring master"

  echo "Configuring the archive directory"

  mkdir -p $PGDATA/archive
  chown -R postgres $PGDATA



  #TODO, we need to do an nslookup here
  allowed_replicas="10.244.0.0/16"

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
synchronous_standby_names = '$SYNCHONROUS_REPLICAS'
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

  echo "Beginning bootstrap from host $MASTER_ENDPOINT"

  gosu postgres pg_basebackup -D $PGDATA -p 5432 -U postgres -v -h $MASTER_ENDPOINT --xlog-method=stream

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
  primary_conninfo = 'host=$MASTER_ENDPOINT port=5432 user=postgres application_name=$SYNCHONROUS_REPLICA'
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
  echo "FAIL!  I need built"
}


#get the short hostname for comparison
#
# HOSTNAME="foo-1"
# FQDN="foo-1.bar.baz.local"


echo "Configuring postres"
echo "Hostname is $HOSTNAME"
echo "FQDN is $(hostname -f)"
echo "MEMBER_ROLE is $MEMBER_ROLE"
#Make sure the postgres process owns the mounted directory



if [[ "$MEMBER_ROLE" == "master" ]]; then
  configure_master
elif [[ "$MEMBER_ROLE" == "slave"  ]]; then
  configure_slave
else
  configure_replica
fi
