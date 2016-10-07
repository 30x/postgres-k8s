#!/bin/bash

#Assumes a synchrnous replica so data is not lost.  See section 25.2.8.3. Planning for High Availability
#https://www.postgresql.org/docs/9.5/static/warm-standby.html#SYNCHRONOUS-REPLICATION


# Get the first synchrnous replica.  Make that the master, then create another slave to replace it

#Exec into the first online slave and make it master

if [ "$1" == "" ]; then
  echo "You must specify a cluster name"
  exit 1
fi

#Get only healthy running pods for the cluster that are a slave
REMAINING_PODS=$(kubectl get po |grep "postgres-$1-" |grep "Running"|awk '{print $1}'|sort)

if [ "$REMAINING_PODS" == "" ]; then
  echo "No pods could be found"
  exit 1
fi

if [ "$(echo $REMAINING_PODS |grep postgres-$1-0 |wc -l)" -eq 1 ]; then
  echo "The master already appears running. Short circuting"
  exit 1
fi


# for pod in $SLAVE_PODS; do
#   echo "Found pod $pod as a failover candidate"
# done

select POD in $REMAINING_PODS;
do
     echo "You picked $POD ($REPLY), making this the master."
     break
done


#Logging into the pod doesn't work.  We need to update it to be configured as the master
#with the current

# Get the tags based on the pod name
#replica: "SLAVE_INDEX"

PERSISTENT_VOLUME_CLAIM=$(kubectl get po $POD -o='jsonpath={$.spec.volumes[:1].persistentVolumeClaim.claimName}')

if [ "$PERSISTENT_VOLUME_CLAIM" == "" ]; then
  echo "Could not find the persistent volume claim to set into the master. Exiting"
  exit 1
fi


CLUSTER_NAME=$(kubectl get po $POD -o='jsonpath={$.metadata.labels.cluster}')

if [ "$CLUSTER_NAME" == "" ]; then
  echo "Could not find the cluster name to set into the master. Exiting"
  exit 1
fi

INDEX=$(kubectl get po $POD -o='jsonpath={$.metadata.labels.index}')

if [ "$INDEX" == "" ]; then
  echo "Could not find the node index to set into the master. Exiting"
  exit 1
fi

#TODO Derive this from either 1) user input or 2) system state.  1) will most likely be safer
SYNCHRNOUS_REPLICAS="1,2,3,4,5"

#Create a temporary directory to process files
TEMP_DIR=$(mktemp -d)

FILENAME="$TEMP_DIR/pg-master-rs.yaml"

echo "Configuring new Replica Set to replace slave in file $FILENAME"

#Here we replace the whole template with an exact match. This way if someone has changed the name, it will still work
sed "s/pg-data-CLUSER_NAME_TO_REPLACE-DISK_INDEX/$PERSISTENT_VOLUME_CLAIM/g" kubernetes/pg-failover-rs.yaml > $FILENAME

sed -i '' "s/CLUSER_NAME_TO_REPLACE/$CLUSTER_NAME/g" $FILENAME

sed -i '' "s/SLAVE_NAMES/$SYNCHRNOUS_REPLICAS/g" $FILENAME
#Put the disk index, this will create an immutable PVC to attach
sed -i '' "s/DISK_INDEX/$INDEX/g" $FILENAME

#Now we have a functoning master template, replace the existing rs with this one
kubectl replace -f $FILENAME
kubectl delete po $POD
