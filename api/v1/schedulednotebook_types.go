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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ScheduledNotebookSpec defines the desired state of ScheduledNotebook
type ScheduledNotebookSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The name of the notebook workflows
	Name string `json:"name,omitempty"`
	// The docker image to use to execute the notebooks (must contain papermill and required libs)
	DockerImage string `json:"dockerImage"`
	// CronLike syntax for schedule
	// See https://en.wikipedia.org/wiki/Cron
	Schedule string `json:"schedule"`
	// URI of input notebook
	InputNotebook string `json:"inputNotebook"`
	// URI of output notebook
	OutputNotebook string `json:"outputNotebook"`
	// Key value pairs for notebook parameters
	Parameters map[string]string `json:"parameters"`
}

// ScheduledNotebookStatus defines the observed state of ScheduledNotebook
type ScheduledNotebookStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The number of actively running pods
	Active int32 `json:"active"`
	// Represents time when the job was completed. It is not guaranteed to
	// be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	CompletionTime metav1.Time `json:"completionTime"`
	// The latest available observations of an object's current state.
	// More info: https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/
	Conditions []metav1.Condition `json:"conditions"`
	// The number of pods which reached phase Failed.
	Failed int32 `json:"failed"`
	// Represents time when the job was acknowledged by the job controller.
	// It is not guaranteed to be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	StartTime metav1.Time `json:"startTime"`
	// The number of pods which reached phase Succeeded.
	Succeeded int32 `json:"succeeded"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ScheduledNotebook is the Schema for the schedulednotebooks API
type ScheduledNotebook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScheduledNotebookSpec   `json:"spec,omitempty"`
	Status ScheduledNotebookStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ScheduledNotebookList contains a list of ScheduledNotebook
type ScheduledNotebookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScheduledNotebook `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ScheduledNotebook{}, &ScheduledNotebookList{})
}
