/*
Copyright 2021 wanglet.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KubesprayNode struct {
	Name     string            `json:"name"`
	Role     string            `json:"role"`
	Host     string            `json:"host,omitempty"`
	User     string            `json:"user,omitempty"`
	Password string            `json:"password,omitempty"`
	Vars     map[string]string `json:"vars,omitempty"`
}

type KubesprayJobType string

var (
	KubesprayJobTypeScale               KubesprayJobType = "scale"
	KubesprayJobTypeRemoveNode          KubesprayJobType = "remove-node"
	KubesprayJobTypeUpgrade             KubesprayJobType = "upgrade"
	KubesprayJobTypeRecoverControlPlane KubesprayJobType = "recover-control-plane"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KubesprayJobSpec defines the desired state of KubesprayJob
type KubesprayJobSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Type               KubesprayJobType `json:"type"`
	Nodes              []KubesprayNode  `json:"nodes"`
	ExtraVarsConfigmap string           `json:"extraVarsConfigmap"`
}

// KubesprayJobStatus defines the observed state of KubesprayJob
type KubesprayJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase          corev1.PodPhase `json:"phase,omitempty"`
	StartTime      *metav1.Time    `json:"startTime,omitempty"`
	CompletionTime *metav1.Time    `json:"completionTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KubesprayJob is the Schema for the kubesprayjobs API
type KubesprayJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubesprayJobSpec   `json:"spec,omitempty"`
	Status KubesprayJobStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KubesprayJobList contains a list of KubesprayJob
type KubesprayJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubesprayJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubesprayJob{}, &KubesprayJobList{})
}
