
# Overview

The point of this script is to configure postgres in k8s for a performance test.
The system is automated with manual event firing. Eventually, this will be encapsulated
in an api with a sidecar monitor that will perform these steps automatically

## Starting a cluster

Bootstrap your kuberetes service to have dynamic AWS type allocated.  Once this has been performed on a K8s cluster, it will not need performed again
```
kubectl create -f kuberetes/storageclass.yaml
```

Now create your first cluster.  We'll call it "test"

```
./admin-scripts/createcluster.sh test
```

This will spin up a cluster.  Make a note of the read and write service names that were provided on the startup.

You can monitor the pod status by performing the following

```
kubectl get po -l app=postgres,cluster=test
```

Note that you may observe errors until the master has started.  Once the master is successfully running, all pods should eventually go to a "Running" STATUS
You can now connect to the PG service names that were returned from the create cluster command.  In a clean fresh cluster, pod with index 0 is always the master.
You can search for which pod is currently the master with the following kubectl command.

```
kubectl get po -l app=postgres,cluster=test,master=true
```

## Testing Pod Failure

In the event a pod dies, it will be restarted automatically and the cluster will normalize, no intervention should be required.  You can test this by killing any one pod with the command below

```
kubectl delete po {pod name}
```

Performing a select of the pods with a get will show the pods restarting

```
kubectl get po -l app=postgres,cluster=test
```

## Testing Availability Zone Failure

In the event an AZ with the master pod dies, a failover to a slave is required.  To perform a failover, you must manually run the failover command.  
Eventually this will be an automated process that fails over to a slave automatically in the event the master pod terminates and cannot be recovered.

Simulate a pod death by completely removing the master's Replication Service(RS).  This will halt all streaming to the slaves.
```
kubectl get po -l app=postgres,cluster=test,master=true
```

Then take the suffix, and delete the RS with the following command.  This assumed the pod was of the format `postgres-test-0-{some id}`.

```
kubectl delete rs postgres-test-0
```

You will now see both slaves are unable to connect to the master by using log --follow on the pod.  Now, failover to a new slave node.

```
./admin-scripts/failover.sh test
```

Pick a node to fail over to. After the execution of the kubectl you will notice the pod you have selected is the master, the service points to it, and replication is restored.  At this point, you should use the addreplica.sh script to scale out to return your cluster to 1 mater and synchrnous replicas.  You can verify the master is up and running by selecting the master pod.

```
kubectl get po -l app=postgres,cluster=test,master=true
```

## Adding Read Capacity

TODO

## Tearing down the cluster

To remove the cluster, but retain the persistent volumes to resume operation, you would perform the following.

```
./admin-scripts/deletecluster.sh  test
```

You could then resume the cluster state with the pods by simply creating the cluster again.  If the Persistent Volume Claims (PVC) exist, they will be reused.

```
./admin-scripts/createcluster.sh test
```

To completely remove the cluster, and delete all data, you should run the following command

```
./admin-scripts/deletecluster.sh -d test
```

Note the '-d' flag to delete the PVC.  Use with great care.  Once the PVC has been deleted, the EBS will be deleted as well

## Notes on the gory details of synchronous replication that aren't well documented

1. The application_name parameter in the slave's recovery.conf must be a string match
to the name specified in the masters synchronous_standby_names parameter in postgrsql.conf
