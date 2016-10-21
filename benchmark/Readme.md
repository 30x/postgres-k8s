# Postgres pg_bench tests

The purpose of this docker image is to allow a user to schedule a job within k8s to perform stress testing of any PG cluster.
It's original intent is to test our PG+transiactor

# Running

The goal of the project is to be easily executable.  It makes the following assumptions.

1. You have a functioning Kubernetes cluster >= 1.4.0 and have the ability to create jobs
1. You have access to create an AWS S3 bucket and upload objects

Start by creating a secret with the AWS creds that will be used to upload results to S3. Make sure `kubernetes` is your current working directory.
You will only need to set this up the first time you run the test, or when your AWS creds change.

```
./createsecret.sh -k [YOUR KEY] -s [YOUR SECRET] -p [POSTGRES PASSWORD] -n [YOUR NAMESPACE]
```
