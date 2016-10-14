#!/bin/sh


#This executes a sql script.  The script creates a database, inserts a row, then selects a sum.  Afterwards, the DB is removed
gosu postgres psql << EOF
\x
CREATE DATABASE test_failover;
\c test_failover;
CREATE TABLE test (
    value         integer
);

/*insert the value of 10*/
INSERT INTO test (value) VALUES (10);

/*Select the sum, we don't want to use 1, since this appears in the output*/
SELECT SUM(value) from TEST;

/*Connect to postgres to drop the test database*/
\c postgres;
DROP DATABASE test_failover;
EOF
