#!/bin/bash

# -----
# Boostraps an instance into k8s.  Assumes the kubectl is already pointing at the
# correct env
# ----

if [ $1 -eq ""]; then
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

  FILENAME="$TEMP_DIR/pg-slave$index.yaml"

  sed "s/CLUSER_NAME_TO_REPLACE/$CLUSTER_NAME/g" kubernetes/pg-slave.yaml > $FILENAME
  sed -i '' "s/SLAVE_INDEX/$index/g" $FILENAME

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

FILENAME="$TEMP_DIR/pg-master.yaml"

sed "s/CLUSER_NAME_TO_REPLACE/$CLUSTER_NAME/g" kubernetes/pg-master.yaml > $FILENAME

sed -i '' "s/SLAVE_INDEX/$SYNCHRNOUS_REPLICAS/g" $FILENAME

#Now start master
kubectl create -f $FILENAME

echo "Created master node"

echo ""
echo ""
echo "##########"
echo ""

echo "Creation complete"
echo "Write service endpoint is $CLUSTER_NAME-write"
echo "Read service endpoint is $CLUSTER_NAME-read"
