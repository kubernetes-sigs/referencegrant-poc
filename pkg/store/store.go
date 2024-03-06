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

package store

import (
	"sync"

	authorizationv1 "k8s.io/api/authorization/v1"
)

// AuthorizationStore interface
type AuthorizationStore interface {
	CheckAuthz(sar authorizationv1.SubjectAccessReview) (bool, error)
	// AddGrant(resource, verb string, allowed bool)
}

// in-memory GraphStore
type AuthStore struct {
	graph map[string]map[string]bool
	mutex sync.RWMutex
}

func NewGraphStore() *AuthStore {
	return &AuthStore{graph: make(map[string]map[string]bool)}
}

func (s *AuthStore) CheckAuthz(sar authorizationv1.SubjectAccessReview) (bool, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	resource := sar.Spec.ResourceAttributes.Resource
	verb := sar.Spec.ResourceAttributes.Verb

	if verbs, ok := s.graph[resource]; ok {
		if allowed, ok := verbs[verb]; ok {
			return allowed, nil
		}
	}
	return false, nil
}

func (s *AuthStore) AddGrant(resource, verb string, allowed bool) {
	if _, ok := s.graph[resource]; !ok {
		s.graph[resource] = make(map[string]bool)
	}
	s.graph[resource][verb] = allowed
}
