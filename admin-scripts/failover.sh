#!/bin/bash

#Assumes a synchrnous replica so data is not lost.  See section 25.2.8.3. Planning for High Availability
#https://www.postgresql.org/docs/9.5/static/warm-standby.html#SYNCHRONOUS-REPLICATION


# Get the first synchrnous replica.  Make that the master, then create another slave to replace it

#Exec into the first online slave and make it master

if [ "$1" == "" ]; then
  echo "You must specify a cluster name"
  exit 1
fi

MASTER_POD=$(kubectl get po --no-headers -l "app=postgres,cluster=$1,master=true"|awk '{print $1}')

if [ "$MASTER_POD" != "" ]; then
  echo "The master already appears running in pod $MASTER_POD. Short circuting"
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


#Get the new master's index, so that we skip it when listing the rest
NEW_MASTER_INDEX=$(kubectl get po $POD -o='jsonpath={$.metadata.labels.index}')


#GET ALL running system's index
ALL_INDEXES=$(kubectl get po -l "app=postgres,cluster=$1" -o='jsonpath={$.items[*].metadata.labels.index}')

#The list of new slaves to construct
NEW_SLAVES=""

for index in $ALL_INDEXES; do
  #SKIP the new master index
  if [ "$index" == "$NEW_MASTER_INDEX" ]; then
    continue
  fi

  if [ "$NEW_SLAVES" == "" ]; then
    NEW_SLAVES="${index}"
  else
    NEW_SLAVES="${NEW_SLAVES},${index}"
  fi

done


echo "New slaves to set into master PG is $NEW_SLAVES"

#Create the touch file to promote the server to master
kubectl  exec $POD bash /clusterutils/setreplicas.sh $NEW_SLAVES

echo "Signaling to pod $POD to become active master"

kubectl exec -ti $POD /usr/bin/touch /tmp/postgresql.trigger.5432

echo "Labeling the new pod as the master"

kubectl label pods $POD master=true

echo "Testing the new master.  If this takes longer than 30 seconds, the failover may not have worked.  Verify manually"
#Add the label so that the write/replication service endpoint picks up the new service
SUMRESULT=$(kubectl  exec $POD bash /clusterutils/testdb.sh | grep sum | grep 10 |wc -l)

if [ $SUMRESULT -eq 1 ]; then
  echo "SUCCESS: Master is successfully running and replicating to other replicas"
else
  echo "FAILURE: Could not create a test database on the master. Something is incorrect with the current replication setup"
fi
