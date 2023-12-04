/*
Copyright 2023.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ConsulKVSpec defines the desired state of ConsulKV
type ConsulKVSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	QoS       QoSType `json:"qos,omitempty"`
	ConsulUrl string  `json:"consul_url,omitempty"`

	Paths []PathSpec `json:"paths,omitempty"`

	GuardAgainst     []string `json:"guard_against,omitempty"`
	WhitelistedPaths []string `json:"whitelisted_paths,omitempty"`
}

type PathSpec struct {
	// +kubebuilder:validation:MinLength=1
	Path string `json:"path,omitempty"`

	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	CriticalityWeight int `json:"criticality_weight"`
}

type QoSType string

var (
	Relaxed  QoSType = "relaxed"
	Medium   QoSType = "medium"
	Critical QoSType = "critical"
)

type AdaptationMode string

var (
	NonAdaptive    AdaptationMode = "non-adaptive"
	SelfHealing    AdaptationMode = "self-healing"
	SelfProtecting AdaptationMode = "self-protecting"
)

// ConsulKVStatus defines the observed state of ConsulKV
type ConsulKVStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	UtilityFunctionValue string         `json:"utility_function_value"`
	AdaptationMode       AdaptationMode `json:"adaptation_mode"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ConsulKV is the Schema for the consulkvs API
type ConsulKV struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConsulKVSpec   `json:"spec,omitempty"`
	Status ConsulKVStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ConsulKVList contains a list of ConsulKV
type ConsulKVList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConsulKV `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ConsulKV{}, &ConsulKVList{})
}
