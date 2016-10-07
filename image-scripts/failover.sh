#!/bin/bash

#invoke the init script to reconfigure the slave
bash /docker-entrypoint-initdb.d/init.sh

exec gosu postgres postgres
