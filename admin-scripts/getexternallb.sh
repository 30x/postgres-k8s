#!/bin/bash

#Assumes a synchrnous replica so data is not lost.  See section 25.2.8.3. Planning for High Availability
#https://www.postgresql.org/docs/9.5/static/warm-standby.html#SYNCHRONOUS-REPLICATION


# Get the first synchrnous replica.  Make that the master, then create another slave to replace it

#Exec into the first online slave and make it master

if [ "$1" == "" ]; then
  echo "You must specify a cluster name"
  exit 1
fi

WRITE_ENDPOINT=$(kubectl describe svc  -l "app=postgres,type=write,cluster=$1"|grep "LoadBalancer Ingress")

if [ "$WRITE_ENDPOINT" == "" ]; then
  echo "Could not find the write service. Short circuting"
  exit 1
fi


READ_ENDPOINT=$(kubectl describe svc  -l "app=postgres,type=read,cluster=$1"|grep "LoadBalancer Ingress")

if [ "$READ_ENDPOINT" == "" ]; then
  echo "Could not find the read service. Short circuting"
  exit 1
fi

echo "Write endpoint for postgres is $WRITE_ENDPOINT"
echo "Read endpoint for postgres is $READ_ENDPOINT"
