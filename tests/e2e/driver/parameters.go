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

// SizeForVolumeType returns the minimum disk size for each volumeType
func SizeForVolumeType(volumeType string) string {
	switch volumeType {
	case "st1":
		return "500Gi"
	case "sc1":
		return "500Gi"
	case "gp2":
		return "1Gi"
	case "io1":
		return "4Gi"
	default:
		return "1Gi"
	}
}

// IOPSPerGBForVolumeType returns the minimum value for io1 volumeType
// Otherwise returns an empty string
func IOPSPerGBForVolumeType(volumeType string) string {
	if volumeType == "io1" {
		return "3"
	}
	return ""
}
