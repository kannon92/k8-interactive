/*
Copyright 2022.

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

package v1

import (
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IngressConfig struct {
	Default            bool              `json:"default,omitempty"`
	Ports              []int32           `json:"ports,omitempty"`
	IngressAnnotations map[string]string `json:"ingressannotations,omitempty"`
	TlsEnabled         bool              `json:"tlsenabled,omitempty"`
	CertName           string            `json:"certname,omitempty"`
}

type ServiceConfig struct {
	Ports []int32 `json:"ports"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// InteractiveJobSpec defines the desired state of InteractiveJob
type InteractiveJobSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of InteractiveJob. Edit interactivejob_types.go to remove/update
	JobTemplate batchv1.JobSpec `json:"jobtemplate"`
	Ingress     IngressConfig   `json:"ingress,omitempty"`
	Service     ServiceConfig   `json:"service"`
}

// +kubebuilder:printcolumn:name="JobStatus",type=string,JSONPath=`.status.jobstatus`
// InteractiveJobStatus defines the observed state of InteractiveJob
type InteractiveJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	JobStatus string `json:"jobstatus,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// InteractiveJob is the Schema for the interactivejobs API
type InteractiveJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InteractiveJobSpec   `json:"spec,omitempty"`
	Status InteractiveJobStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InteractiveJobList contains a list of InteractiveJob
type InteractiveJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InteractiveJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InteractiveJob{}, &InteractiveJobList{})
}
