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

package main

import (
	"fmt"
	"github.com/kubernetes-sigs/aws-ebs-csi-driver/tests/e2e/driver"
	"github.com/kubernetes-sigs/aws-ebs-csi-driver/tests/e2e/testsuites"
	"k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/storage/drivers"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("[ebs-csi] Dynamic Provisioning", func() {

	f := framework.NewDefaultFramework("ebs")

	var (
		cs clientset.Interface
		ns *v1.Namespace
	)

	BeforeEach(func() {
		cs = f.ClientSet
		ns = f.Namespace
		config := framework.VolumeTestConfig{
			Namespace: ns.Name,
			Prefix:    "ebs",
		}
		drivers.SetCommonDriverParameters(ebsDriver, f, config)
	})

	for _, t := range driver.ValidVolumeTypes {
		for _, fs := range driver.ValidFSTypes {
			volumeType := t
			fsType := fs
			Context(fmt.Sprintf("with %q volumeType and %q fsType", volumeType, fsType), func() {
				It("should create a volume on demand", func() {
					// Generate StorageClass, PVC and testing reading and writing
					parameters := map[string]string{
						"type":   volumeType,
						"fsType": fsType,
					}
					if iops := driver.IOPSPerGBForVolumeType(volumeType); iops != "" {
						parameters["iopsPerGB"] = iops
					}
					storageClass := ebsDriver.GetDynamicProvisionStorageClass(parameters, v1.PersistentVolumeReclaimDelete, ns.Name)
					claimSize := driver.SizeForVolumeType(volumeType)
					pvClaim := newClaim(storageClass.Name, claimSize, ns.Name)
					scTest := testsuites.StorageClassTest{
						StorageClass:          storageClass,
						PersistentVolumeClaim: pvClaim,
						Client:                cs,
					}
					testsuites.TestDynamicProvisioning(scTest)
				})
			})
		}
	}
})
