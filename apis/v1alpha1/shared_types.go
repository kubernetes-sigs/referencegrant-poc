/*
Copyright 2024 The Kubernetes Authors.

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

import "fmt"

type GroupResource struct {
	Group    string `json:"group"`
	Resource string `json:"resource"`
}

type GroupResourceNamespace struct {
	Group     string `json:"group"`
	Resource  string `json:"resource"`
	Namespace string `json:"namespace"`
}

type For string

type QueueKey struct {
	OriginGroup    string
	OriginResource string
	TargetGroup    string
	TargetResource string
	Purpose        string
}

func (qk *QueueKey) ToString() string {
	origin := fmt.Sprintf("%s/%s", qk.OriginGroup, qk.OriginResource)
	target := fmt.Sprintf("%s/%s", qk.TargetGroup, qk.TargetResource)
	return fmt.Sprintf("%s;%s;%s", origin, target, qk.Purpose)
}

func NewQueueKey(originGroup, originResource, targetGroup, targetResource, purpose string) *QueueKey {
	return &QueueKey{
		OriginGroup:    originGroup,
		OriginResource: originResource,
		TargetGroup:    targetGroup,
		TargetResource: targetResource,
		Purpose:        purpose,
	}
}
