#!/bin/bash

#Assumes a synchrnous replica so data is not lost.  See section 25.2.8.3. Planning for High Availability
#https://www.postgresql.org/docs/9.5/static/warm-standby.html#SYNCHRONOUS-REPLICATION


# Get the first synchrnous replica.  Make that the master, then create another slave to replace it
