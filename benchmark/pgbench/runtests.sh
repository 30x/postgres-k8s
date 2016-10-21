#!/bin/bash

if [ "$PG_HOST" == "" ]; then
  echo "You must set the PG_HOST env var"
  exit 1
fi


if [ "$PG_USER" == "" ]; then
  echo "You must set the PG_USER env var"
  exit 1
fi


if [ "$PG_PORT" == "" ]; then
  echo "You must set the PG_PORT env var"
  exit 1
fi


if [ "$PG_PASSWORD" == "" ]; then
  echo "You must set the PG_PASSWORD env var"
  exit 1
fi

if [ "$AWS_ACCESS_KEY_ID" == "" ]; then
  echo "You must set the AWS_ACCESS_KEY_ID env var"
  exit 1
fi


if [ "$AWS_SECRET_ACCESS_KEY" == "" ]; then
  echo "You must set the AWS_SECRET_ACCESS_KEY env var"
  exit 1
fi


if [ "$S3_BUCKET" == "" ]; then
  echo "You must set the S3_BUCKET env var"
  exit 1
fi

if [ "$S3_REGION" == "" ]; then
  echo "You must set the S3_REGION env var"
  exit 1
fi


workingdir=$(pwd)

echo "Working directory is $workingdir"

echo "Verifying you have S3 access to the bucket $S3_BUCKET in region $S3_REGION"


OUTPUT=$(aws s3api head-bucket --region $S3_REGION --bucket $S3_BUCKET 2>&1)

if [ $? -ne 0 ]; then

  missing=$(echo $OUTPUT|grep 404|wc -l)

  if [ $missing -eq 1 ]; then
    echo "Bucket $S3_BUCKET does not exist in region $S3_REGION.  Creating"
    OUTPUT=$(aws s3api create-bucket --region $S3_REGION --bucket $S3_BUCKET 2>&1)

    if [ $? -ne 0 ]; then
      echo "Unable to create bucket"
      echo $OUTPUT
      exit 1
    fi
  else
    echo "An error occured when trying to setup the AWS bucket."
    echo $OUTPUT
    exit 1
  fi
fi

#Set our password so that any psql/pg client picks it up
export PGPASSWORD=$PG_PASSWORD

#Source the config file
. ./config

#Now create the tables
echo "Creating the results database"

createdb -U $PG_USER -h $PG_HOST -p $PG_PORT results

echo "Creating the pgbench database"

createdb -U $PG_USER -h $PG_HOST -p $PG_PORT pgbench

echo "Initializing the pgbench database"
pgbench  -U $PG_USER -h $PG_HOST -p $PG_PORT  -i

echo "Initializing the results database"
psql -U $PG_USER -h $PG_HOST -p $PG_PORT -f init/resultdb.sql -d results


echo "Running tests"

./runset

echo "Copying results"


#If we get here, we can access the bucket, send up the results
TIME=$(date +%s)
DIRNAME="$PG_HOST/$TIME"
S3_NAME="$DIRNAME/results.tar.gz"

#Tar up results
tar cvzf results.tar.gz results

aws s3api put-object --region $S3_REGION --bucket $S3_BUCKET --key $S3_NAME --body results.tar.gz

echo "done" > done.txt

S3_NAME="$DIRNAME/done.txt"

aws s3api put-object --region $S3_REGION --bucket $S3_BUCKET --key $S3_NAME --body done.txt
