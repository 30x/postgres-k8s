#!/bin/bash


if [ "$1" == "" ]; then
  echo "You must specify a replica string"
  exit 1
fi


NEW_SLAVES=$1


TIMESTAMP=(date +%s)

PG_FILE="$PGDATA/postgresql.conf"
BACKUP_FILE="$PG_FILE.$TIMESTAMP"

#Copy the file to the backup
cp $PG_FILE $BACKUP_FILE

#Overwrite PG with the new file and the new slaves
sed -e "s/synchronous_standby_names.*/synchronous_standby_names='$NEW_SLAVES'/g"  $BACKUP_FILE > $PG_FILE

#Set ownership of files
chown -R postgres $PGDATA
