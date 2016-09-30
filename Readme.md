
# Overview

The point of this script is to configure postgres in k8s for a performance test.

## Steps

1. Create 2 100GB Gp2 EBS volumes in the EC2 console.  Not the volume ids
1. Modify the file pg-pvmaster.yaml and pg-pvslave.yaml to contain the persistent volume ids from the previous step.
1. Create the pets
```
  kubectl create -f pg-pets.yaml
```


## Notes on synchronous replication that aren't well documented

1. The application_name parameter in the slave's recovery.conf must be a string match
to the name specified in the masters synchronous_standby_names parameter in postgrsql.conf
