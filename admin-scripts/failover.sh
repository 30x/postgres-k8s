#!/bin/bash

#Assumes a synchrnous replica so data is not lost.  See section 25.2.8.3. Planning for High Availability
#https://www.postgresql.org/docs/9.5/static/warm-standby.html#SYNCHRONOUS-REPLICATION


# Get the first synchrnous replica.  Make that the master, then create another slave to replace it

#Exec into the first online slave and make it master

if [ "$1" == "" ]; then
  echo "You must specify a cluster name"
  exit 1
fi

MASTER_POD=$(kubectl get po --no-headers -l "app=postgres,cluster=$1,master=true"|wc -l)

if [ "$MASTER_POD" -eq 1 ]; then
  echo "The master already appears running. Short circuting"
  exit 1
fi


#Get only healthy running pods for the cluster that are a slave
REMAINING_PODS=$(kubectl get po --no-headers -l "app=postgres,cluster=$1"|grep "Running"|awk '{print $1}'|sort)

if [ "$REMAINING_PODS" == "" ]; then
  echo "No pods could be found"
  exit 1
fi




#Allow the user to select a pod
select POD in $REMAINING_PODS;
do
     echo "You picked $POD ($REPLY), making this the master."
     break
done



#Create the touch file to promote the server to master
kubectl exec -ti $POD /usr/bin/touch /tmp/postgresql.trigger.5432

#Add the label so that the write/replication service endpoint picks up the new service
kubectl label pods $POD master=true

#TODO add sanity checks here for failover to ensure it's functioning correctly

echo "If you have an old master running, you may now safely remove"
