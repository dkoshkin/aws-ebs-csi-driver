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

package testsuites

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	imageutils "k8s.io/kubernetes/test/utils/image"
	"time"
)

// StorageClassTest represents parameters to be used by CSI tests
type StorageClassTest struct {
	Name                  string
	StorageClass          *storagev1.StorageClass
	PersistentVolumeClaim *v1.PersistentVolumeClaim
	SkipWriteReadCheck    bool
	Client                clientset.Interface
}

// TestDynamicProvisioning tests dynamic provisioning with specified StorageClassTest
// Provisions the StorageClass
// Provisions the PersistentVolumeClaim
// Runs a pod and write and read some data
func TestDynamicProvisioning(t StorageClassTest) {
	storageClass, storageClassCleanup, err := ProvisionStorageClass(t.StorageClass, t.Client)
	if storageClassCleanup != nil {
		defer storageClassCleanup()
	}
	Expect(err).NotTo(HaveOccurred())

	persistedVolumeClaim, persistedVolumeClaimCleanup, err := ProvisionPersistentVolumeClaim(t.PersistentVolumeClaim, t.Client)
	if persistedVolumeClaimCleanup != nil {
		defer persistedVolumeClaimCleanup()
	}
	Expect(err).NotTo(HaveOccurred())

	persistedVolume := ValidateProvisionedPersistentVolume(storageClass, t.PersistentVolumeClaim, persistedVolumeClaim, t.Client)

	if !t.SkipWriteReadCheck {
		podCleanup, err := ValidatePodCanWriteAndRead(persistedVolumeClaim, t.Client)
		if podCleanup != nil {
			podCleanup()
		}
		Expect(err).NotTo(HaveOccurred())
	}

	// Wait for the PV to get deleted if reclaim policy is Delete. (If it's
	// Retain, there's no use waiting because the PV won't be auto-deleted and
	// it's expected for the caller to do it.) Technically, the first few delete
	// attempts may fail, as the volume is still attached to a node because
	// kubelet is slowly cleaning up the previous pod, however it should succeed
	// in a couple of minutes.
	if persistedVolume.Spec.PersistentVolumeReclaimPolicy == v1.PersistentVolumeReclaimDelete {
		By(fmt.Sprintf("deleting the claim's PV %q", persistedVolume.Name))
		framework.ExpectNoError(framework.WaitForPersistentVolumeDeleted(t.Client, persistedVolume.Name, 5*time.Second, 10*time.Minute))
	}

}

func ProvisionStorageClass(storageClass *storagev1.StorageClass, client clientset.Interface) (*storagev1.StorageClass, func(), error) {
	By("creating a StorageClass " + storageClass.Name)
	class, err := client.StorageV1().StorageClasses().Create(storageClass)
	cleanup := func() {
		framework.Logf("deleting storage class %s", class.Name)
		framework.ExpectNoError(client.StorageV1().StorageClasses().Delete(class.Name, nil))
	}
	return class, cleanup, err
}

func ProvisionPersistentVolumeClaim(persistentVolumeClaim *v1.PersistentVolumeClaim, client clientset.Interface) (*v1.PersistentVolumeClaim, func(), error) {
	By("creating a PersistentVolumeClaim")
	claim, err := client.CoreV1().PersistentVolumeClaims(persistentVolumeClaim.Namespace).Create(persistentVolumeClaim)
	Expect(err).NotTo(HaveOccurred())
	cleanup := func() {
		framework.Logf("deleting PersistentVolumeClaim %q/%q", claim.Namespace, claim.Name)
		// typically this claim has already been deleted
		err = client.CoreV1().PersistentVolumeClaims(claim.Namespace).Delete(claim.Name, nil)
		if err != nil && !apierrs.IsNotFound(err) {
			framework.Failf("Error deleting claim %q. Error: %v", claim.Name, err)
		}
	}
	err = framework.WaitForPersistentVolumeClaimPhase(v1.ClaimBound, client, claim.Namespace, claim.Name, framework.Poll, framework.ClaimProvisionTimeout)
	if err != nil {
		return nil, cleanup, err
	}

	By("checking the PersistentVolumeClaim")
	// Get new copy of the claim
	claim, err = client.CoreV1().PersistentVolumeClaims(claim.Namespace).Get(claim.Name, metav1.GetOptions{})
	return claim, cleanup, err

}

func ValidateProvisionedPersistentVolume(storageClass *storagev1.StorageClass, requestedPersistentVolumeClaim *v1.PersistentVolumeClaim, persistentVolumeClaim *v1.PersistentVolumeClaim, client clientset.Interface) *v1.PersistentVolume {
	// Get the bound PersistentVolume
	persistentVolume, err := client.CoreV1().PersistentVolumes().Get(persistentVolumeClaim.Spec.VolumeName, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())

	// Check sizes
	expectedCapacity := requestedPersistentVolumeClaim.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	claimCapacity := persistentVolumeClaim.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	Expect(claimCapacity.Value()).To(Equal(expectedCapacity.Value()), "claimCapacity is not equal to requestedCapacity")

	pvCapacity := persistentVolume.Spec.Capacity[v1.ResourceName(v1.ResourceStorage)]
	Expect(pvCapacity.Value()).To(Equal(expectedCapacity.Value()), "pvCapacity is not equal to requestedCapacity")

	// Check PV properties
	By("checking the PersistentVolume")
	expectedAccessModes := requestedPersistentVolumeClaim.Spec.AccessModes
	Expect(persistentVolume.Spec.AccessModes).To(Equal(expectedAccessModes))
	Expect(persistentVolume.Spec.ClaimRef.Name).To(Equal(persistentVolumeClaim.ObjectMeta.Name))
	Expect(persistentVolume.Spec.ClaimRef.Namespace).To(Equal(persistentVolumeClaim.ObjectMeta.Namespace))
	Expect(persistentVolume.Spec.PersistentVolumeReclaimPolicy).To(Equal(*storageClass.ReclaimPolicy))
	Expect(persistentVolume.Spec.MountOptions).To(Equal(storageClass.MountOptions))

	return persistentVolume
}

func ValidatePodCanWriteAndRead(persistentVolumeClaim *v1.PersistentVolumeClaim, client clientset.Interface) (func(), error) {
	By("checking a pod can write data and then read the same data from a volume")
	command := "echo 'hello world' > /mnt/test/data && grep 'hello world' /mnt/test/data"
	return runInPodWithVolume(persistentVolumeClaim, command, client)
}

// runInPodWithVolume runs a command in a pod with given claim mounted to /mnt directory.
func runInPodWithVolume(persistentVolumeClaim *v1.PersistentVolumeClaim, command string, c clientset.Interface) (func(), error) {
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "pvc-volume-tester-",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:    "volume-tester",
					Image:   imageutils.GetE2EImage(imageutils.BusyBox),
					Command: []string{"/bin/sh"},
					Args:    []string{"-c", command},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "my-volume",
							MountPath: "/mnt/test",
						},
					},
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
			Volumes: []v1.Volume{
				{
					Name: "my-volume",
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: persistentVolumeClaim.Name,
							ReadOnly:  false,
						},
					},
				},
			},
		},
	}

	ns := persistentVolumeClaim.Namespace
	pod, err := c.CoreV1().Pods(ns).Create(pod)
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %v", err)
	}

	framework.ExpectNoError(err, "Failed to create pod: %v", err)
	cleanup := func() {
		framework.Logf("deleting Pod %q/%q", pod.Name, pod.Namespace)
		body, err := c.CoreV1().Pods(ns).GetLogs(pod.Name, &v1.PodLogOptions{}).Do().Raw()
		if err != nil {
			framework.Logf("Error getting logs for pod %s: %v", pod.Name, err)
		} else {
			framework.Logf("Pod %s has the following logs: %s", pod.Name, body)
		}
		framework.DeletePodOrFail(c, ns, pod.Name)
	}
	return cleanup, framework.WaitForPodSuccessInNamespaceSlow(c, pod.Name, pod.Namespace)
}
