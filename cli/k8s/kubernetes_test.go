package k8s_test

import (
	"fmt"
	"math"
	"strconv"

	"github.com/30x/postgres-k8s/cli/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/pkg/api/v1"
)

var _ = Describe("kubernetes", func() {

	It("Storage Class", func() {
		name := "testStorageClass"
		storageClass := k8s.CreateStorageClass(name)

		Expect(storageClass.Name).Should(Equal(name))

		Expect(storageClass.Provisioner).Should(Equal("kubernetes.io/aws-ebs"))

		Expect(storageClass.Parameters["type"]).Should(Equal("gp2"))

	})

	It("PersistentVolumeClaim", func() {
		name := "testStorageClass"
		class := "postgrestestclass"
		index := 1
		sizeInGigs := 250

		pvc := k8s.CreatePersistentVolumeClaim(name, class, index, sizeInGigs)

		expectedName := fmt.Sprintf("pg-data-%s-%d", name, index)

		Expect(pvc.Name).Should(Equal(expectedName))
		Expect(pvc.Annotations["volume.beta.kubernetes.io/storage-class"]).Should(Equal(class))
		Expect(pvc.Labels["app"]).Should(Equal("postgres"))
		Expect(pvc.Labels["cluster"]).Should(Equal(name))
		Expect(pvc.Spec.AccessModes[0]).Should(Equal(v1.ReadWriteOnce))

		storageResource := pvc.Spec.Resources.Requests[v1.ResourceStorage]
		outputSize := storageResource.Value()
		expectedSize := int64(sizeInGigs) * int64(math.Pow(1024, 3))

		Expect(outputSize).Should(Equal(expectedSize))
	})

	It("Master Replica Set", func() {
		clusterName := "testCluster"
		slaveNames := []string{"foo", "bar"}

		// sizeInGigs := 250

		rs := k8s.CreateMaster(clusterName, slaveNames, 0)

		expectedName := fmt.Sprintf("postgres-%s-%d", clusterName, 0)

		Expect(rs.ObjectMeta.Name).Should(Equal(expectedName))
		Expect(rs.Spec.Template.Labels["app"]).Should(Equal("postgres"))
		Expect(rs.Spec.Template.Labels["cluster"]).Should(Equal(clusterName))
		Expect(rs.Spec.Template.Labels["index"]).Should(Equal("0"))
		Expect(rs.Spec.Template.Labels["role"]).Should(Equal("master"))
		Expect(rs.Spec.Template.Labels["master"]).Should(Equal("true"))

		container := &rs.Spec.Template.Spec.Containers[0]

		Expect(container.Name).Should(Equal("postgres"))

		Expect(container.Image).Should(Equal(k8s.Image))

		env := container.Env[0]
		Expect(env.Name).Should(Equal("POSTGRES_PASSOWRD"))
		Expect(env.Value).Should(Equal("password"))

		env = container.Env[1]
		Expect(env.Name).Should(Equal("PGDATA"))
		Expect(env.Value).Should(Equal("/pgdata/data"))

		env = container.Env[2]
		Expect(env.Name).Should(Equal("PGMOUNT"))
		Expect(env.Value).Should(Equal("/pgdata"))

		env = container.Env[3]
		Expect(env.Name).Should(Equal("MEMBER_ROLE"))
		Expect(env.Value).Should(Equal("master"))

		env = container.Env[4]
		Expect(env.Name).Should(Equal("SYNCHONROUS_REPLICAS"))
		Expect(env.Value).Should(Equal("foo,bar"))

		env = container.Env[5]
		Expect(env.Name).Should(Equal("WAL_LEVEL"))
		Expect(env.Value).Should(Equal("logical"))
		Expect(container.Ports[0].Name).Should(Equal("postgres"))
		Expect(container.Ports[0].ContainerPort).Should(Equal(int32(5432)))

	})

	It("Replica Replica Set", func() {
		clusterName := "testCluster"
		index := 1

		// sizeInGigs := 250

		rs := k8s.CreateReplica(clusterName, index)

		expectedName := fmt.Sprintf("postgres-%s-%d", clusterName, 1)

		Expect(rs.ObjectMeta.Name).Should(Equal(expectedName))
		Expect(rs.Spec.Template.Labels["app"]).Should(Equal("postgres"))
		Expect(rs.Spec.Template.Labels["cluster"]).Should(Equal(clusterName))
		Expect(rs.Spec.Template.Labels["index"]).Should(Equal(strconv.Itoa(index)))
		Expect(rs.Spec.Template.Labels["role"]).Should(Equal("replica"))

		container := &rs.Spec.Template.Spec.Containers[0]

		Expect(container.Name).Should(Equal("postgres"))

		Expect(container.Image).Should(Equal(k8s.Image))

		env := container.Env[0]
		Expect(env.Name).Should(Equal("POSTGRES_PASSOWRD"))
		Expect(env.Value).Should(Equal("password"))

		env = container.Env[1]
		Expect(env.Name).Should(Equal("PGDATA"))
		Expect(env.Value).Should(Equal("/pgdata/data"))

		env = container.Env[2]
		Expect(env.Name).Should(Equal("PGMOUNT"))
		Expect(env.Value).Should(Equal("/pgdata"))

		env = container.Env[3]
		Expect(env.Name).Should(Equal("MEMBER_ROLE"))
		Expect(env.Value).Should(Equal("slave"))

		expectedName = fmt.Sprintf("postgres-%s-write", clusterName)

		env = container.Env[4]
		Expect(env.Name).Should(Equal("MASTER_ENDPOINT"))
		Expect(env.Value).Should(Equal(expectedName))

		env = container.Env[5]
		Expect(env.Name).Should(Equal("SYNCHONROUS_REPLICA"))
		Expect(env.Value).Should(Equal(strconv.Itoa(index)))

		Expect(container.Ports[0].Name).Should(Equal("postgres"))
		Expect(container.Ports[0].ContainerPort).Should(Equal(int32(5432)))

	})

	It("Write Service", func() {
		clusterName := "testCluster"

		// sizeInGigs := 250

		service := k8s.CreateWriteService(clusterName)

		expectedName := fmt.Sprintf("postgres-%s-write", clusterName)

		Expect(service.Name).Should(Equal(expectedName))

		Expect(service.Spec.Selector["app"]).Should(Equal("postgres"))
		Expect(service.Spec.Selector["master"]).Should(Equal("true"))
		Expect(service.Spec.Selector["cluster"]).Should(Equal(clusterName))

	})

	It("Read Service", func() {
		clusterName := "testCluster"

		service := k8s.CreateReadService(clusterName)

		expectedName := fmt.Sprintf("postgres-%s-read", clusterName)

		Expect(service.Name).Should(Equal(expectedName))

		Expect(service.Spec.Selector["app"]).Should(Equal("postgres"))
		Expect(service.Spec.Selector["role"]).Should(Equal("replica"))
		Expect(service.Spec.Selector["cluster"]).Should(Equal(clusterName))
	})
})
