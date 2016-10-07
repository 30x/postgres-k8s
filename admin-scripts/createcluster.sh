#!/bin/bash

# -----
# Boostraps an instance into k8s.  Assumes the kubectl is already pointing at the
# correct env.
#
# This script is idempotent with respect to complete resource creation.  In the event a service is accidentally deleted, such
# as a replica set, or a service, this script can be re run with the same cluster name to re-create those resources.
#
# Note that if the master persisatent volume is accidentally deleted, the failover script should be run to make a slave the master, then the slave will be re-created
# ----

if [ $1 == ""]; then
  echo "You must specify a cluster name"
  exit 1
fi

NUM_SLAVES="2"

echo "Args are $@"

if [  ! -z $2]; then
  NUM_SLAVES=$2
fi

CLUSTER_NAME=$1

#Create a temporary directory to process files
TEMP_DIR=$(mktemp -d)

echo "Placing yaml files into directory $TEMP_DIR"
echo "Creating a cluster wtih 1 master and $NUM_SLAVES synchronous slaves"

echo "Creating services"

sed s/CLUSER_NAME_TO_REPLACE/$CLUSTER_NAME/g kubernetes/pg-services.yaml > $TEMP_DIR/pg-services.yaml

kubectl create -f $TEMP_DIR/pg-services.yaml


#Loop through slaves and create them here
SYNCHRNOUS_REPLICAS=""

for index in $(seq 1 ${NUM_SLAVES}); do
  # SLAVE_NAME="slave-$index"

  echo "Creating slave ${SLAVE_NAME}"

  FILENAME="$TEMP_DIR/pg-slave-rs-$index.yaml"

  sed "s/CLUSER_NAME_TO_REPLACE/$CLUSTER_NAME/g" kubernetes/pg-slave-rs.yaml > $FILENAME
  #Put the SLAVE_INDEX this is immutable once a pod is running
  sed -i '' "s/SLAVE_INDEX/$index/g" $FILENAME
  #Put the disk index, this will create an immutable PVC to attach
  sed -i '' "s/DISK_INDEX/$index/g" $FILENAME


  if [ $index -eq 1 ]; then
    SYNCHRNOUS_REPLICAS="${index}"
  else
    SYNCHRNOUS_REPLICAS="${SYNCHRNOUS_REPLICAS},${index}"
  fi

  #Now create the slave in kubernetes
  kubectl create -f $FILENAME

  echo "Created slave ${index}"
done

echo "Creating master node"
echo "SYNCHRNOUS_REPLICAS are $SYNCHRNOUS_REPLICAS"

FILENAME="$TEMP_DIR/pg-master-rs.yaml"

sed "s/CLUSER_NAME_TO_REPLACE/$CLUSTER_NAME/g" kubernetes/pg-master-rs.yaml > $FILENAME

sed -i '' "s/SLAVE_NAMES/$SYNCHRNOUS_REPLICAS/g" $FILENAME
#Put the disk index, this will create an immutable PVC to attach
sed -i '' "s/DISK_INDEX/0/g" $FILENAME

#Now start master
kubectl create -f $FILENAME

echo "Created master node"

echo ""
echo ""
echo "##########"
echo ""

echo "Creation complete"
echo "Write service endpoint is postgres-$CLUSTER_NAME-write"
echo "Read service endpoint is postgres-$CLUSTER_NAME-read"
