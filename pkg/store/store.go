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
	"fmt"
	"log"
	"strings"
	"sync"

	authorizationv1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	v1a1 "sigs.k8s.io/referencegrant-poc/apis/v1alpha1"
)

const ExtraPurposeKey = "purpose"

// AuthorizationStore interface
type AuthorizationStore interface {
	CheckAuthz(sar authorizationv1.SubjectAccessReview) (bool, error)
	// AddGrant(resource, verb string, allowed bool)
}

// currently "group/resource"
type TargetResourceGroup string

// currently "ns/name"
// type TargetNamespacedName string

// This is the "For" string
type Purpose string

// Initial version of the graph - maps "from-to-for" to a map of target("to")resource names to set of subjects
// This ignores classNames, namespace, verbs.
type GrantGraph map[string]map[types.NamespacedName]sets.Set[v1a1.Subject]

// in-memory AuthStore
type AuthStore struct {
	graph GrantGraph
	// TODO: should we support multiple purposes for the same subject ?
	subjectIndex map[v1a1.Subject]map[TargetResourceGroup]map[types.NamespacedName]Purpose
	mutex        sync.RWMutex
}

func (s *AuthStore) GetGraph() GrantGraph {
	return s.graph
}

func (s *AuthStore) GetSubjectIndex() map[v1a1.Subject]map[TargetResourceGroup]map[types.NamespacedName]Purpose {
	return s.subjectIndex
}

func NewAuthStore() *AuthStore {
	return &AuthStore{
		graph:        make(GrantGraph),
		subjectIndex: make(map[v1a1.Subject]map[TargetResourceGroup]map[types.NamespacedName]Purpose),
		mutex:        sync.RWMutex{},
	}
}

func (s *AuthStore) CheckAuthz(sar authorizationv1.SubjectAccessReview) (bool, error) {
	user := sar.Spec.User
	groups := sar.Spec.Groups
	trg := TargetResourceGroup(fmt.Sprintf("%s/%s", sar.Spec.ResourceAttributes.Group, sar.Spec.ResourceAttributes.Resource))

	nn := types.NamespacedName{
		Name:      sar.Spec.ResourceAttributes.Name,
		Namespace: sar.Spec.ResourceAttributes.Namespace,
	}

	// Not necessary for the demo.
	// var p Purpose
	// if _, ok := sar.Spec.Extra[ExtraPurposeKey]; ok {
	// 	p = Purpose(strings.Join(sar.Spec.Extra[ExtraPurposeKey], ", "))
	// }

	s.mutex.RLock()
	defer s.mutex.RUnlock()
	for _, g := range groups {
		allowed := s.lookup(v1a1.Subject{Kind: "Group", Name: g}, trg, nn)
		if allowed {
			return true, nil
		}
	}
	allowed := s.lookup(v1a1.Subject{Kind: "User", Name: user}, trg, nn)
	return allowed, nil
}

// Graph Lookup is perfomed as follows:
//
//  1. Start by attempting to find the subject (subj) in the authorization graph subjectIndex.
//     If it's not found, it returns false indicating that the subject doesn't have any grants.
//
//  2. It then looks for the target resource group (trg) within the subject's map in the graph.
//     If not found, it returns false.
//
//  3. Next, it looks for the namespacedName within the target resource group's map.
//     if not found, it returns false.
//     If found, it checks if the found Purpose is the requested purpose and return true or false accordingly.
func (s *AuthStore) lookup(subj v1a1.Subject, trg TargetResourceGroup, nn types.NamespacedName) bool {
	if nn.Name == "tls-validity-checks-certificate" {
		// log.Printf("Attempting to lookup graph for with subj=%v, trg=%v, nn=%v, purpose=%v", subj, trg, nn, p)
		log.Printf("Attempting to lookup graph for with subj=%v, trg=%v, nn=%v", subj, trg, nn)
		log.Printf("SubjectIndex is: %v", s.GetSubjectIndex())
	}
	trgMap, ok := s.subjectIndex[subj]
	if !ok {
		return false
	}
	tnnMap, ok := trgMap[trg]
	if !ok {
		return false
	}
	// We dont care what is the purpose it is authorized as we currently have no way to get the "Purpose" or "For" from the client
	_, ok = tnnMap[nn]
	return ok

}

func (s *AuthStore) UpsertGrant(key string, resourceName types.NamespacedName, subjects []v1a1.Subject) {
	splitedKey := strings.Split(key, ";")
	to, purpose := TargetResourceGroup(splitedKey[1]), Purpose(splitedKey[2])
	// resourceNameFmt := TargetNamespacedName(resourceName)

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if _, ok := s.graph[key]; !ok {
		s.graph[key] = make(map[types.NamespacedName]sets.Set[v1a1.Subject])
	}
	if _, ok := s.graph[key][resourceName]; !ok {
		s.graph[key][resourceName] = make(sets.Set[v1a1.Subject])
	}
	s.graph[key][resourceName].Insert(subjects...)

	for _, subject := range subjects {
		if _, ok := s.subjectIndex[subject]; !ok {
			s.subjectIndex[subject] = make(map[TargetResourceGroup]map[types.NamespacedName]Purpose)
		}
		if _, ok := s.subjectIndex[subject][to]; !ok {
			s.subjectIndex[subject][to] = make(map[types.NamespacedName]Purpose)
		}
		s.subjectIndex[subject][to][resourceName] = purpose
	}
}

func (s *AuthStore) ClearGraphKey(key string) {
	splitedKey := strings.Split(key, ";")
	to, purpose := TargetResourceGroup(splitedKey[1]), Purpose(splitedKey[2])

	s.mutex.RLock()
	defer s.mutex.RUnlock()
	delete(s.graph, key)
	// Update subject index
	for subject, targetGroupResourceMap := range s.subjectIndex {
		for tgr, tnnMap := range targetGroupResourceMap {
			if tgr != to {
				continue
			}
			for tnn, p := range tnnMap {
				if p == purpose {
					// If the key is found in the subjectIndex, remove the entry
					delete(s.subjectIndex[subject][tgr], tnn)
				}
			}
			// If the map under a particular To becomes empty after deletion,
			// remove the entire To entry from the subjectIndex
			if len(s.subjectIndex[subject][tgr]) == 0 {
				delete(s.subjectIndex[subject], tgr)
			}
		}
		// If the map under a particular subject becomes empty after deletion,
		// remove the entire subject entry from the subjectIndex
		if len(s.subjectIndex[subject]) == 0 {
			delete(s.subjectIndex, subject)
		}
	}
}
