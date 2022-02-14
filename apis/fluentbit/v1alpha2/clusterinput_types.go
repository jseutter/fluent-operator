/*
Copyright 2021.

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

package v1alpha2

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"

	"fluent.io/fluent-operator/apis/fluentbit/v1alpha2/plugins"
	"fluent.io/fluent-operator/apis/fluentbit/v1alpha2/plugins/input"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// InputSpec defines the desired state of ClusterInput
type InputSpec struct {
	// A user friendly alias name for this input plugin.
	// Used in metrics for distinction of each configured input.
	Alias string `json:"alias,omitempty"`
	// Dummy defines Dummy Input configuration.
	Dummy *input.Dummy `json:"dummy,omitempty"`
	// Tail defines Tail Input configuration.
	Tail *input.Tail `json:"tail,omitempty"`
	// Systemd defines Systemd Input configuration.
	Systemd *input.Systemd `json:"systemd,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +genclient

// ClusterInput is the Schema for the inputs API
type ClusterInput struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec InputSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:generate:=false
// InputByName implements sort.Interface for []Input based on the Name field.
type InputByName []ClusterInput

func (a InputByName) Len() int           { return len(a) }
func (a InputByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a InputByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

// +kubebuilder:object:root=true

// ClusterInputList contains a list of Input
type InputList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterInput `json:"items"`
}

func (list InputList) Load(sl plugins.SecretLoader) (string, error) {
	var buf bytes.Buffer

	sort.Sort(InputByName(list.Items))

	for _, item := range list.Items {
		merge := func(p plugins.Plugin) error {
			if p == nil || reflect.ValueOf(p).IsNil() {
				return nil
			}

			buf.WriteString("[Input]\n")
			buf.WriteString(fmt.Sprintf("    Name    %s\n", p.Name()))
			if item.Spec.Alias != "" {
				buf.WriteString(fmt.Sprintf("    Alias    %s\n", item.Spec.Alias))
			}
			kvs, err := p.Params(sl)
			if err != nil {
				return err
			}
			buf.WriteString(kvs.String())
			return nil
		}

		for i := 0; i < reflect.ValueOf(item.Spec).NumField(); i++ {
			p, _ := reflect.ValueOf(item.Spec).Field(i).Interface().(plugins.Plugin)
			if err := merge(p); err != nil {
				return "", err
			}
		}
	}

	return buf.String(), nil
}

func init() {
	SchemeBuilder.Register(&ClusterInput{}, &InputList{})
}