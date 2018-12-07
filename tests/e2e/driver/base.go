/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
   http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package driver

import (
	"k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/kubernetes/test/e2e/storage/drivers"
)

// DynamicPVTestDriver represents an interface for a TestDriver that supports DynamicPV
type DynamicPVTestDriver interface {
	drivers.TestDriver
	// GetDynamicProvisionStorageClass returns a StorageClass dynamic provision Persistent Volume.
	GetDynamicProvisionStorageClass(parameters map[string]string, reclaimPolicy v1.PersistentVolumeReclaimPolicy, namespace string) *storagev1.StorageClass
}
