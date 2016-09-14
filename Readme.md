
# Overview

The point of this script is to configure postgres in k8s for a performance test.

## Steps

1. Create 2 100GB Gp2 EBS volumes in the EC2 console.  Not the volume ids
1. Modify the file pg-pvmaster.yaml and pg-pvslave.yaml to contain the persistent volume ids from the previous step.
1. Create the pets
```
  kubectl create -f pg-pets.yaml
```

## Manual configuration

This will ultimately become obcelete with scripting.  For now, these are the manual steps.

### Configure replication on the master

Ultimately this will be scripted


```
kubectl exec -ti edgexpostgres-0 -- bash

#Create the user repuser with 5 connections and grants replication
sudo -u postgres createuser -U postgres repuser  -c 5 --replication

sudo -u postgres psql -c "ALTER USER repuser WITH ENCRYPTED PASSWORD 'testreplicationpassword'"
```

Now create the test backup file.  TODO should this even be on the EBS?

```
mkdir -p /var/lib/postgresql/data/archive
chown postgres /var/lib/postgresql/data/archive
```


Now configure the ha file to allow replication


```
cat <<EOF >> /var/lib/postgresql/data/pgdata/pg_hba.conf
host     replication     repuser        192.168.0.0/32        md5
EOF
```

Lastly configure the WAL files


```
cat <<EOF >> /var/lib/postgresql/data/pgdata/pg_hba.conf

wal_level = hot_standby
archive_mode = on
archive_command = 'test ! -f /var/lib/postgresql/data/archive/%f && cp %p /var/lib/postgresql/data/archive/%f'
max_wal_senders = 3

EOF
```


Restore backup to Slave

```MASTER={INSERT HOSTNAME/IP}

su -c "/usr/pgsql-9.4/bin/pg_ctl -D /ebs/pgdata stop -m fast" postgres
rm -rf /ebs/pgdata/*
chown -R postgres: /ebs/pgdata

su -c "/usr/pgsql-9.4/bin/pg_basebackup -D /ebs/pgdata -p 5432 -U postgres -v -h $MASTER --xlog-method=stream" postgres

su -c "/usr/pgsql-9.4/bin/pg_ctl -D /ebs/pgdata start" postgres

#test
psql -U postgres -d apigee -c "SELECT now() - pg_last_xact_replay_timestamp() AS time_lag";
```

Edit slave's pg_hba.conf to allow replica, then repeat steps above on replica


Lastly edit postgresql.conf and add the following parameters to enable synchronous replication

```
synchronous_standby_names = '{slave ip}'
```

Notes in testing

Test client

m4.4xlarge
