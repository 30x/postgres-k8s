package k8s

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/resource"
	"k8s.io/client-go/pkg/api/v1"
	extv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	apiv1beta1 "k8s.io/client-go/pkg/apis/storage/v1beta1"
	"k8s.io/client-go/pkg/util/intstr"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

//Image the image we're pulling. Eventually this should come from a var that's overwritten by the build process, or a local configuration yaml file.
const Image = "thirtyx/postgres:0.0.3-dev"

//CreateClientFromEnv create a k8s client from the current runtime.  Searches
func CreateClientFromEnv() (*kubernetes.Clientset, error) {
	//try the local configuration first
	clientSet, err := createLocalConfig()

	if err == nil {
		return clientSet, nil
	}

	//log it
	fmt.Println(err)

	fmt.Printf("Unable to find user configuration.  Falling back to in cluster configuration")

	return createInClusterConfig()
}

//Create the client set from the in cluster configuration
func createInClusterConfig() (*kubernetes.Clientset, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()

	if err != nil {
		return nil, err
	}
	// creates the clientset
	return kubernetes.NewForConfig(config)
}

func createLocalConfig() (*kubernetes.Clientset, error) {

	//load up the kuberentes default file location
	//TODO, maybe make this an env var?

	home := os.Getenv("HOME")

	if home == "" {
		return nil, errors.New("Could not get the HOME directory env variable")
	}

	kubeconfig := filepath.Join(home, ".kube", "config")

	exists, err := exists(kubeconfig)

	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("%s does not exist", kubeconfig)
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)

	if err != nil {
		return nil, err
	}

	// creates the clientset
	return kubernetes.NewForConfig(config)

}

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

//CreatePersistentVolumeClaim create a pvc with the information provided
func CreatePersistentVolumeClaim(clusterName, storageClass string, diskIndex, sizeInGigs int) *v1.PersistentVolumeClaim {

	/*
	   kind: PersistentVolumeClaim
	   apiVersion: v1
	   metadata:
	     name: pg-data-CLUSER_NAME_TO_REPLACE-DISK_INDEX
	     annotations:
	       volume.beta.kubernetes.io/storage-class: postgresv1
	     labels:
	       app: postgres
	       cluster: "CLUSER_NAME_TO_REPLACE"
	       index: "DISK_INDEX"
	   spec:
	     accessModes:
	       - ReadWriteOnce
	     resources:
	       requests:
	         storage: 100Gi
	*/
	name := getPersistentVolumeClaimName(clusterName, diskIndex)
	size := getQuantityInGigs(sizeInGigs)

	pvc := &v1.PersistentVolumeClaim{}

	pvc.Name = name
	pvc.Annotations = make(map[string]string)
	pvc.Annotations["volume.beta.kubernetes.io/storage-class"] = storageClass

	pvc.Labels = make(map[string]string)
	pvc.Labels["app"] = "postgres"
	pvc.Labels["cluster"] = clusterName
	pvc.Labels["index"] = strconv.Itoa(diskIndex)

	pvc.Spec.AccessModes = []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce}

	pvc.Spec.Resources.Requests = make(v1.ResourceList)
	pvc.Spec.Resources.Requests[v1.ResourceStorage] = size

	return pvc
}

//CreateStorageClass for now it's AWS only
func CreateStorageClass(storageClassName string) *apiv1beta1.StorageClass {

	/*
	   kind: StorageClass
	   apiVersion: storage.k8s.io/v1beta1
	   metadata:
	     name: postgresv1
	   provisioner: kubernetes.io/aws-ebs
	   parameters:
	     type: gp2

	*/

	storageClass := &apiv1beta1.StorageClass{}

	storageClass.Name = storageClassName

	storageClass.Provisioner = "kubernetes.io/aws-ebs"

	storageClass.Parameters = make(map[string]string)
	storageClass.Parameters["type"] = "gp2"

	return storageClass

	// storageClass.

}

//CreateMaster create the master for the cluster
func CreateMaster(clusterName string, replicaIDs []string, index int) *extv1beta1.ReplicaSet {
	/**
	    apiVersion: extensions/v1beta1
	kind: ReplicaSet
	metadata:
	  name: postgres-CLUSER_NAME_TO_REPLACE-DISK_INDEX
	spec:
	  # Always 1 replica, we have a different RC per PG instance
	  replicas: 1
	  template:
	    metadata:
	      labels:
	        app: postgres
	        role: master
	        cluster: "CLUSER_NAME_TO_REPLACE"
	        index: "DISK_INDEX"
	    spec:
	      terminationGracePeriodSeconds: 0
	      containers:
	      - name: postgres
	        image: thirtyx/postgres:0.0.3-dev
	        env:
	          - name: POSTGRES_PASSWORD
	            value: password
	          - name: PGDATA
	            value: /pgdata/data
	          - name: PGMOUNT
	            value: /pgdata
	          - name: MEMBER_ROLE
	            value: master
	          - name: SYNCHONROUS_REPLICAS
	            value: "SLAVE_NAMES"
	          - name: WAL_LEVEL
	            value: logical
	        ports:
	          - containerPort: 5432
	        volumeMounts:
	        - mountPath: /pgdata
	          name:  pg-data-CLUSER_NAME_TO_REPLACE-DISK_INDEX
	        imagePullPolicy: Always
	      volumes:
	      - name: pg-data-CLUSER_NAME_TO_REPLACE-DISK_INDEX
	        persistentVolumeClaim:
	          claimName: pg-data-CLUSER_NAME_TO_REPLACE-DISK_INDEX

	---
	kind: PersistentVolumeClaim
	apiVersion: v1
	metadata:
	  name: pg-data-CLUSER_NAME_TO_REPLACE-DISK_INDEX
	  annotations:
	    volume.beta.kubernetes.io/storage-class: postgresv1
	  labels:
	    app: postgres
	    cluster: "CLUSER_NAME_TO_REPLACE"
	    index: "DISK_INDEX"
	spec:
	  accessModes:
	    - ReadWriteOnce
	  resources:
	    requests:
	      storage: 100Gi
	      **/
	rs := createBaseReplicaSet(clusterName, index)

	rs.Spec.Template.Labels["role"] = "master"

	container := &rs.Spec.Template.Spec.Containers[0]

	var buffer bytes.Buffer

	for i, inputError := range replicaIDs {

		if i > 0 {
			buffer.WriteString(",")
		}

		buffer.WriteString(inputError)

	}

	replicaString := buffer.String()

	//append the master env vars
	appendEnv(container, "MEMBER_ROLE", "master")
	appendEnv(container, "SYNCHONROUS_REPLICAS", replicaString)
	appendEnv(container, "WAL_LEVEL", "logical")

	return rs

}

//CreateReplica create a replica
func CreateReplica(clusterName string, index int) *extv1beta1.ReplicaSet {
	/*	apiVersion: extensions/v1beta1
		kind: ReplicaSet
		metadata:
		  name: postgres-CLUSER_NAME_TO_REPLACE-DISK_INDEX
		spec:
		  # Always 1 replica, we have a different RC per PG instance
		  replicas: 1
		  template:
		    metadata:
		      labels:
		        app: postgres
		        role: replica
		        cluster: "CLUSER_NAME_TO_REPLACE"
		        index: "DISK_INDEX"
		    spec:
		      terminationGracePeriodSeconds: 0
		      containers:
		      - name: postgres
		        image: thirtyx/postgres:0.0.3-dev
		        env:
		          - name: POSTGRES_PASSWORD
		            value: password
		          - name: PGDATA
		            value: /pgdata/data
		          - name: PGMOUNT
		            value: /pgdata
		          - name: MEMBER_ROLE
		            value: slave
		          - name: MASTER_ENDPOINT
		            value: postgres-CLUSER_NAME_TO_REPLACE-write
		            #The name of the synchronous replica.  This will need to be included
		            # in the string for the variable of the master node
		          - name: SYNCHONROUS_REPLICA
		            value: "SLAVE_INDEX"
		        ports:
		          - containerPort: 5432
		        volumeMounts:
		        - mountPath: /pgdata
		          name:  pg-data-CLUSER_NAME_TO_REPLACE-DISK_INDEX
		        imagePullPolicy: Always
		      volumes:
		      - name: pg-data-CLUSER_NAME_TO_REPLACE-DISK_INDEX
		        persistentVolumeClaim:
		          claimName: pg-data-CLUSER_NAME_TO_REPLACE-DISK_INDEX

	*/
	rs := createBaseReplicaSet(clusterName, index)

	rs.Spec.Template.Labels["role"] = "replica"

	masterService := getWriteServiceName(clusterName)

	container := &rs.Spec.Template.Spec.Containers[0]

	replicaID := strconv.Itoa(index)

	//append the master env vars
	appendEnv(container, "MEMBER_ROLE", "slave")
	appendEnv(container, "MASTER_ENDPOINT", masterService)
	appendEnv(container, "SYNCHONROUS_REPLICA", replicaID)

	return rs

}

func createBaseReplicaSet(clusterName string, index int) *extv1beta1.ReplicaSet {
	/*apiVersion: extensions/v1beta1
	kind: ReplicaSet
	metadata:
	  name: postgres-CLUSER_NAME_TO_REPLACE-DISK_INDEX
	spec:
	  # Always 1 replica, we have a different RC per PG instance
	  replicas: 1
	  template:
	    metadata:
	      labels:
	        app: postgres
	        role: slave
	        cluster: "CLUSER_NAME_TO_REPLACE"
	        read: "true"
	        index: "DISK_INDEX"
	    spec:
	      terminationGracePeriodSeconds: 0
	      containers:
	      - name: postgres
	        image: thirtyx/postgres:0.0.3-dev
	        env:
	          - name: POSTGRES_PASSWORD
	            value: password
	          - name: PGDATA
	            value: /pgdata/data
	          - name: PGMOUNT
	            value: /pgdata
	          - name: MEMBER_ROLE
	            value: slave
	          - name: MASTER_ENDPOINT
	            value: postgres-CLUSER_NAME_TO_REPLACE-write
	            #The name of the synchronous replica.  This will need to be included
	            # in the string for the variable of the master node
	          - name: SYNCHONROUS_REPLICA
	            value: "SLAVE_INDEX"
	        ports:
	          - containerPort: 5432
	        volumeMounts:
	        - mountPath: /pgdata
	          name:  pg-data-CLUSER_NAME_TO_REPLACE-DISK_INDEX
	        imagePullPolicy: Always
	      volumes:
	      - name: pg-data-CLUSER_NAME_TO_REPLACE-DISK_INDEX
	        persistentVolumeClaim:
	          claimName: pg-data-CLUSER_NAME_TO_REPLACE-DISK_INDEX
	*/

	rs := &extv1beta1.ReplicaSet{}

	rs.Name = getNodeName(clusterName, index)

	replicas := int32(1)
	rs.Spec.Replicas = &replicas

	rs.Spec.Template.ObjectMeta.Labels = make(map[string]string)

	labels := rs.Spec.Template.ObjectMeta.Labels
	labels["app"] = "postgres"
	labels["cluster"] = clusterName
	labels["index"] = strconv.Itoa(index)

	//point the podspec
	podSpec := &rs.Spec.Template.Spec

	terminationGrace := int64(0)
	podSpec.TerminationGracePeriodSeconds = &terminationGrace
	podSpec.Containers = []v1.Container{v1.Container{}}

	container := &podSpec.Containers[0]

	container.Name = "postgres"
	container.Image = Image
	appendEnv(container, "POSTGRES_PASSOWRD", "password")
	appendEnv(container, "PGDATA", "/pgdata/data")
	appendEnv(container, "PGMOUNT", "/pgdata")

	container.Ports = append(container.Ports, v1.ContainerPort{
		Name:          "postgres",
		ContainerPort: 5432,
	})

	pvcName := getPersistentVolumeClaimName(clusterName, index)

	container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
		MountPath: "/pgdata",
		Name:      pvcName,
	})

	pvcClaim := &v1.PersistentVolumeClaimVolumeSource{
		ClaimName: pvcName,
		ReadOnly:  false,
	}

	//set up the volume
	volume := v1.Volume{}
	volume.Name = pvcName
	volume.PersistentVolumeClaim = pvcClaim

	podSpec.Volumes = append(podSpec.Volumes, volume)

	return rs
}

func appendEnv(container *v1.Container, name, value string) {
	if container.Env == nil {
		container.Env = []v1.EnvVar{}
	}

	envVar := v1.EnvVar{
		Name:  name,
		Value: value,
	}

	container.Env = append(container.Env, envVar)
}

//CreateWriteService create a pvc with the information provided
func CreateWriteService(clusterName string) *v1.Service {

	//TODO, make the dns name and ips configurable
	/*
			  apiVersion: v1
		kind: Service
		metadata:
		  name: postgres-CLUSER_NAME_TO_REPLACE-write
		  labels:
		    app: postgres
		    type: write
		    cluster: "CLUSER_NAME_TO_REPLACE"
		spec:
		  ports:
		  - protocol: TCP
		    port: 5432
		    targetPort: 5432
		    name: postgres
		  # *.edgexpostgres.default.svc.cluster.local
		  selector:
		    app: postgres
		    role: master
		    cluster: "CLUSER_NAME_TO_REPLACE"
		  #Add an external load balancer for read
		  type: LoadBalancer
		  loadBalancerSourceRanges:
		  - 162.246.44.8/32
		  - 50.204.222.32/27
		  - 111.93.155.240/28
		  - 203.145.181.112/28
		  - 52.1.124.126/32
		  - 50.0.116.61/32
		  - 194.74.103.192/28
		  - 50.242.80.16/29
		  - 67.6.206.65/32
		  - 54.200.58.80/32
	*/

	name := getWriteServiceName(clusterName)

	svc := createBaseServiceDef(clusterName)

	svc.Name = name
	svc.ObjectMeta.Labels["type"] = "write"

	svc.Spec.Selector["role"] = "master"

	return svc
}

//CreateReadService create a pvc with the information provided
func CreateReadService(clusterName string) *v1.Service {

	/*
			  apiVersion: v1
		kind: Service
		metadata:
		  name: postgres-CLUSER_NAME_TO_REPLACE-read
		  labels:
		    app: postgres
		    type: read
		    cluster: "CLUSER_NAME_TO_REPLACE"
		spec:
		  ports:
		  - protocol: TCP
		    port: 5432
		    targetPort: 5432
		    name: postgres
		  # *.edgexpostgres.default.svc.cluster.local
		  selector:
		    app: postgres
		    role: replica
		    cluster: "CLUSER_NAME_TO_REPLACE"
		  #Add an external load balancer for read
		  type: LoadBalancer
		  loadBalancerSourceRanges:
		  - 162.246.44.8/32
		  - 50.204.222.32/27
		  - 111.93.155.240/28
		  - 203.145.181.112/28
		  - 52.1.124.126/32
		  - 50.0.116.61/32
		  - 194.74.103.192/28
		  - 50.242.80.16/29
		  - 67.6.206.65/32
		  - 54.200.58.80/32
		---

	*/
	name := getReadServiceName(clusterName)

	svc := createBaseServiceDef(clusterName)

	svc.Name = name
	svc.ObjectMeta.Labels["type"] = "read"

	svc.Spec.Selector["role"] = "replica"

	return svc
}

func createBaseServiceDef(clusterName string) *v1.Service {

	svc := &v1.Service{}

	svc.ObjectMeta.Labels = make(map[string]string)
	svc.ObjectMeta.Labels["app"] = "postgres"
	svc.ObjectMeta.Labels["cluster"] = clusterName

	svc.Spec.Ports = append(svc.Spec.Ports, v1.ServicePort{
		Protocol: v1.ProtocolTCP,
		Port:     5432,
		TargetPort: intstr.IntOrString{
			IntVal: 5432,
		},
		Name: "postgres",
	})

	svc.Spec.Selector = make(map[string]string)
	svc.Spec.Selector["app"] = "postgres"
	svc.Spec.Selector["cluster"] = clusterName

	svc.Spec.Type = v1.ServiceTypeLoadBalancer

	svc.Spec.LoadBalancerSourceRanges = []string{
		"162.246.44.8/32",
		"50.204.222.32/27",
		"111.93.155.240/28",
		"203.145.181.112/28",
		"52.1.124.126/32",
		"50.0.116.61/32",
		"194.74.103.192/28",
		"50.242.80.16/29",
		"67.6.206.65/32",
		"54.200.58.80/32",
	}

	return svc
}

func getQuantityInGigs(sizeInGigs int) resource.Quantity {
	return resource.MustParse(fmt.Sprintf("%dGi", sizeInGigs))
}

func getPersistentVolumeClaimName(clusterName string, index int) string {
	return fmt.Sprintf("pg-data-%s-%d", clusterName, index)
}

func getNodeName(clusterName string, index int) string {
	return fmt.Sprintf("postgres-%s-%d", clusterName, index)
}

func getWriteServiceName(clusterName string) string {
	return fmt.Sprintf("postgres-%s-write", clusterName)
}

func getReadServiceName(clusterName string) string {
	return fmt.Sprintf("postgres-%s-read", clusterName)
}
