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
	"fmt"
	. "github.com/onsi/ginkgo"
	"k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/test/e2e/storage/drivers"
	"k8s.io/kubernetes/test/e2e/storage/testpatterns"
)

const (
	driverName = "ebs.csi.aws.com"
)

var (
	ValidFSTypes     = []string{"", "ext2", "ext3", "ext4"}
	ValidVolumeTypes = []string{"gp2", "io1", "sc1", "st1"}
)

// Implement k8s.io/kubernetes/test/e2e/storage/drivers.DynamicPVTestDriver interface to reuse upstream e2e tests
type ebsCSIDriver struct {
	//cleanup    func()
	driverInfo drivers.DriverInfo
}

// InitEbsCSIDriver returns ebsCSIDriver that implements DynamicPVTestDriver interface
func InitEbsCSIDriver() DynamicPVTestDriver {
	return &ebsCSIDriver{
		driverInfo: drivers.DriverInfo{
			Name:               driverName,
			FeatureTag:         "",
			MaxFileSize:        testpatterns.FileSizeMedium,
			SupportedFsType:    sets.NewString(ValidFSTypes...),
			IsPersistent:       true,
			IsFsGroupSupported: false,
			IsBlockSupported:   true,
		},
	}
}

func (d *ebsCSIDriver) GetDriverInfo() *drivers.DriverInfo {
	return &d.driverInfo
}

func (d *ebsCSIDriver) SkipUnsupportedTest(pattern testpatterns.TestPattern) {
}

func (d *ebsCSIDriver) GetDynamicProvisionStorageClass(parameters map[string]string, reclaimPolicy v1.PersistentVolumeReclaimPolicy, namespace string) *storagev1.StorageClass {
	// TODO don't hardcode when setting up driver in CreateDriver
	//provisioner := drivers.GetUniqueDriverName(d)
	provisioner := d.driverInfo.Name
	name := fmt.Sprintf("%s-%s-sc", namespace, provisioner)

	return getStorageClass(name, provisioner, parameters, reclaimPolicy)
}

func (d *ebsCSIDriver) CreateDriver() {
	By("CreateDriver is unimplemented and expects the driver to already exist")
}

func (d *ebsCSIDriver) CleanupDriver() {
	By("CleanupDriver is unimplemented")
}

func getStorageClass(
	name string,
	provisioner string,
	parameters map[string]string,
	reclaimPolicy v1.PersistentVolumeReclaimPolicy,
) *storagev1.StorageClass {
	return &storagev1.StorageClass{
		TypeMeta: metav1.TypeMeta{
			Kind: "StorageClass",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Provisioner:   provisioner,
		Parameters:    parameters,
		ReclaimPolicy: &reclaimPolicy,
	}
}
