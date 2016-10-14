#!/bin/sh


#This executes a sql script.  The script creates a database, inserts a row, then selects a sum.  Afterwards, the DB is removed
gosu postgres psql << EOF
  SELECT  now() - pg_last_xact_replay_timestamp() AS time_lag;
EOF
