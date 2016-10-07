#!/bin/bash

# -----
# Boostraps an instance into k8s.  Assumes the kubectl is already pointing at the
# correct env
# ----

DELETE=0

function show_help(){
  echo "Usage is $0 -d (optional, to delete persistent volumes.  Use with care) [cluster name]"
}

#Get the optoinal -d
while getopts "hd" opt; do
    case "$opt" in
        h)
            show_help
            exit 0
            ;;
        d)  DELETE=1
            ;;
        f)  output_file=$OPTARG
            ;;
        '?')
            show_help >&2
            exit 1
            ;;
    esac
done

shift "$((OPTIND-1))" # Shift off the options and optional --.

echo "Args are $@"

if [ $1 == "" ]; then
  echo "You must specify a cluster name"
  show_help
  exit 1
fi


REPLICA_SETS=$(kubectl get rs | grep "$1-"|awk '{print $1}')

for rs in $REPLICA_SETS; do
  echo "deleting replica set $rs-"
  kubectl delete rs $rs
done

SERVICES=$(kubectl get svc | grep "$1-"|awk '{print $1}')

for svc in $SERVICES; do
  echo "deleting service $svc"
  kubectl delete svc $svc
done

#TODO delete PVC.  Deliberately left in tact so they can be reused later

if [ $DELETE -eq 1 ]; then
  PVCS=$(kubectl get pvc | grep "$1-"|awk '{print $1}')

  for pvc in $PVCS; do
    echo "deleting pvc $pvc"
    kubectl delete pvc $pvc
  done
fi
