/*
Copyright 2024.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// UCPClusterSpec defines the desired state of UCPCluster
type UCPClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
    // Number of UCP nodes
    Nodes int32 `json:"nodes"`
    // Docker version
    DockerVersion string `json:"dockerVersion"`
    // UCP admin password
    AdminPassword string `json:"adminPassword"`
    // Docker Hub credentials
    DockerHubUsername string `json:"dockerHubUsername"`
    DockerHubPassword string `json:"dockerHubPassword"`
    DockerHubEmail string `json:"dockerHubEmail"`
	// Foo is an example field of UCPCluster. Edit ucpcluster_types.go to remove/update
	Size int32 `json:"size,omitempty"`
}

// UCPClusterStatus defines the observed state of UCPCluster
type UCPClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	 // UCP instance ID
	 InstanceID string `json:"instanceID,omitempty"`
	 // UCP server URL
	 ServerURL string `json:"serverURL,omitempty"`
	 Nodes []string `json:"nodes"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// UCPCluster is the Schema for the ucpclusters API
type UCPCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UCPClusterSpec   `json:"spec,omitempty"`
	Status UCPClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// UCPClusterList contains a list of UCPCluster
type UCPClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UCPCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&UCPCluster{}, &UCPClusterList{})
}

