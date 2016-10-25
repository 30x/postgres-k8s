package k8s_test

import (
	"testing"

	"github.com/30x/postgres-k8s/cli/k8s"
)

func TestStorageClass(t *testing.T) {
	name := "testStorageClass"
	storageClass := k8s.CreateStorageClass(name)

	if storageClass.Name != name {
		t.Errorf("Expected storage class to be '%s'. Instead it was '%s'", name, storageClass.Name)
	}
}
