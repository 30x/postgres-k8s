#!/bin/bash

#Assumes a synchrnous replica so data is not lost.  See section 25.2.8.3. Planning for High Availability
#https://www.postgresql.org/docs/9.5/static/warm-standby.html#SYNCHRONOUS-REPLICATION


# Get the first synchrnous replica.  Make that the master, then create another slave to replace it

#Exec into the first online slave and make it master

if [ "$1" == "" ]; then
  echo "You must specify a cluster name"
  exit 1
fi

CLUSTER_NAME=$1
NUM_REPLICAS="1"

if [ "$2" != "" ]; then
  NUM_REPLICAS=$2
fi

TEMP_DIR=$(mktemp -d)



MASTER_POD=$(kubectl get po --no-headers -l "app=postgres,cluster=$1,master=true"|awk '{print $1}')

if [ "$MASTER_POD" == "" ]; then
  echo "Could not find the master.  Short circuting"
  exit 1
fi

RUNNING_INDEXES=$(kubectl get po -l "app=postgres,role=slave,cluster=$1" -o='jsonpath={$.items[*].metadata.labels.index}')

if [ "$RUNNING_INDEXES" == "" ]; then
  echo "No running replicas could be found.  Use create to create the initial cluster first.  Existing"
  exit 1
fi

MAX_INDEX=0

for index in $RUNNING_INDEXES; do

  if [ $index -gt $MAX_INDEX ]; then
    MAX_INDEX=$index
  fi
done


echo "Highest numbered replica found is $MAX_INDEX"

#Set to the slave indexes we already have running
NEW_SLAVES=$RUNNING_INDEXES

for current_index in $(seq 1 ${NUM_REPLICAS}); do
  #Set our new index
  index=$(($MAX_INDEX + $current_index ))

  echo "Creating slave ${index}"

  FILENAME="$TEMP_DIR/pg-slave-rs-$index.yaml"

  sed "s/CLUSER_NAME_TO_REPLACE/$CLUSTER_NAME/g" kubernetes/pg-slave-rs.yaml > $FILENAME
  #Put the SLAVE_INDEX this is immutable once a pod is running
  sed -i '' "s/SLAVE_INDEX/$index/g" $FILENAME
  #Put the disk index, this will create an immutable PVC to attach
  sed -i '' "s/DISK_INDEX/$index/g" $FILENAME

  NEW_SLAVES="${NEW_SLAVES} $index"

  #Now create the slave in kubernetes
  kubectl create -f $FILENAME

  echo "Created slave ${index}"


done

#Set the replicas into the master, but only do this after they're all caught up
# SYNCHRNOUS_REPLICAS=""
#
# for slave_index in $NEW_SLAVES; do
#   if [ "$SYNCHRNOUS_REPLICAS" == "" ]; then
#     SYNCHRNOUS_REPLICAS="${slave_index}"
#   else
#     SYNCHRNOUS_REPLICAS="${SYNCHRNOUS_REPLICAS},${slave_index}"
#   fi
#
# done
#
# echo "Setting replicas to $SYNCHRNOUS_REPLICAS in master"
#
#
# kubectl  exec $MASTER_POD bash /clusterutils/setreplicas.sh $SYNCHRNOUS_REPLICAS


#TODO, test these on the client?

#
# echo "Testing the master with replicas.  If this takes longer than 30 seconds, the failover may not have worked.  Verify manually"
# #Add the label so that the write/replication service endpoint picks up the new service
# SUMRESULT=$(kubectl  exec $POD bash /clusterutils/testdb.sh | grep sum | grep 10 |wc -l)
#
# if [ $SUMRESULT -eq 1 ]; then
#   echo "SUCCESS: Master is successfully running and replicating to other replicas"
# else
#   echo "FAILURE: Could not create a test database on the master. Something is incorrect with the current replication setup"
# fi
