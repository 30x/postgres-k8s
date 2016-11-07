
# Overview

The point of this script is to configure postgres in k8s for a performance test.
The system is automated with manual event firing. Eventually, this will be encapsulated
in an api with a sidecar monitor that will perform these steps automatically

## Installation.

Installation is quite simple.  Perform the following make command

```
make build-cli
```

This assumes that $GOPATH/bin is in your path.  If it is not, add the following to your path. 

```
export PATH="$PATH:$GOPATH/bin"
```

Also, you will need a valid kubectl configuration in ~/.kube/config.  This assumes that commands such as `kubectl get po` are functioning.

## Starting a cluster

To start a cluster simply specify the namespace and the cluster name.  This assumes the namespace exists.  If it does not, you will need to create the namespace in k8s via `kubectl create NAMESPACE`

```
pgctl create -n NAMESPACE -c CLUSTERNAME
```

This will allocate disks, start the pods, and begin replication between the nodes.  Note that EBS allocation is generally slow, so it may take some time for the pods to come online.

## Failing over a cluster

This is only neccessary in the event the AZ the master pod is running in fails.  If so, you can easily fail over to a heathly replica with the following command.


```
pgctl failover -n NAMESPACE -c CLUSTERNAME
```

## Scaling up read replicas

To scale up read replicas, use the scale command.

```
pgctl scale  -n NAMESPACE -c CLUSTERNAME -r [NUMBER OF REPLICAS]
```

Note that the current implementation doesn't update the WAL configuration on the master.  As a result, you will need to ensure you do not exceed the `max_wal_senders` on the master, which is currently set to 10


## Upgrading to a new docker image

To upgrade to a new docker image, run the folowing command

```
pgctl upgrade -n NAMESPACE -c CLUSTERNAME -i [DOCKER IMAGE NAME]
```

Note that this is meant to be a minor upgrade.  If a major upgrade is required, a scale out with a new version and a failover will be required to minimize downtime.


## Tearing down the cluster

The cluster can be temporarily removed, or permanently removed.  To stop the cluster but retain the disks, run the following command

```
pgctl delete -n NAMESPACE -c CLUSTERNAME
```

Your cluster can then be restarted later using the create command

```
pgctl create -n NAMESPACE -c CLUSTERNAME
```


If you want to permanently delete your data and the EBS storage, add the `-d` parameter

```
pgctl delete -n NAMESPACE -c CLUSTERNAME -d
```

## Notes on the gory details of synchronous replication that aren't well documented

1. The application_name parameter in the slave's recovery.conf must be a string match
to the name specified in the masters synchronous_standby_names parameter in postgrsql.conf
