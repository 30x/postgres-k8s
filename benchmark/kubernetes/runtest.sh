#!/bin/bash

function show_help(){
  echo "Usage is $0 -U [postgres username] -h [postgres host] -p [postgres port] -s [S3 bucket for result] -r [S3 region for bucket]"
}


PG_HOST=""
PG_PORT=""
PG_USER=""
S3_BUCKET=""
S3_REGION=""

#Get the optoinal -d
while getopts "U:h:p:s:r:" opt; do
    case "$opt" in
        U)  PG_USER=$OPTARG
            ;;
        h)  PG_HOST=$OPTARG
            ;;
        p)  PG_PORT=$OPTARG
            ;;
        s)  S3_BUCKET=$OPTARG
            ;;
        r)  S3_REGION=$OPTARG
            ;;

        '?')
            show_help >&2
            exit 1
            ;;
    esac
done

shift "$((OPTIND-1))" # Shift off the options and optional --.


echo $PG_HOST
echo $PG_PORT
echo $PG_USER
echo $S3_BUCKET
echo $S3_REGION

#Validate input
if [ -z "${PG_USER}" ]; then
    show_help
    exit 1
fi

if [ -z "${PG_HOST}" ]; then
    show_help
    exit 1
fi

if [ -z "${PG_PORT}" ]; then
    show_help
    exit 1
fi

if [ -z "${S3_BUCKET}" ]; then
    show_help
    exit 1
fi


if [ -z "${S3_REGION}" ]; then
    show_help
    exit 1
fi

TEMP_DIR=$(mktemp -d)

FILENAME="$TEMP_DIR/test-job.yaml"

TIMESTAMP=$(date +%s)

JOBNAME="postgres-benchmakr-$TIMESTAMP"

sed "s/POSTGRES_USER_TO_REPLACE/$PG_USER/g" test-job.yaml > $FILENAME

#Put the SLAVE_INDEX this is immutable once a pod is running
sed -i '' "s/POSTGRES_PORT_TO_REPLACE/$PG_PORT/g" $FILENAME
#Put the disk index, this will create an immutable PVC to attach
sed -i '' "s/POSTGRES_HOST_TO_REPLACE/$PG_HOST/g" $FILENAME
#S3 bucket to replace
sed -i '' "s/S3_BUCKET_TO_REPLACE/$S3_BUCKET/g" $FILENAME

sed -i '' "s/S3_REGION_TO_REPLACE/$S3_REGION/g" $FILENAME

sed -i '' "s/JOBNAME/$JOBNAME/g" $FILENAME

kubectl create -f $FILENAME

kubectl get po|grep $JOBNAME

RUNNING=0

while [  $RUNNING -lt 1 ]; do
  RUNNING=$(kubectl get po --no-headers -l "job-name=$JOBNAME" |grep Running|wc -l)
  echo "$RUNNING pods running. Waiting for 1 total."
  sleep 3
done

#Now it's done running, follow it's logs
POD_NAME=$(kubectl get po --no-headers -l "job-name=$JOBNAME"  -o=custom-columns=NAME:.metadata.name)

kubectl logs --follow $POD_NAME
#TODO, download results from S3 on completion?
