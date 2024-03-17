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
	"strings"
	"sync"

	authorizationv1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/referencegrant-poc/apis/v1alpha1"
)

const ExtraPurposeKey = "purpose"
const ExtraOriginKey = "origin-id"

const SameNamespaceGrantor = "SameNamespace"

// AuthorizationStore interface
type AuthorizationStore interface {
	CheckAuthz(sar authorizationv1.SubjectAccessReview) (bool, error)
	// AddGrant(resource, verb string, allowed bool)
}

type subject string
type targetResourceGroup schema.GroupResource
type CRPID string
type originResourceID string
type targetNamespacedName struct {
	namespace, name string
}
type purpose string
type verb string

type grant struct {
	verbs    []verb
	grantors []string
}

// grantGraph is structured to allow per-instance insert/update/delete for subject, CRP, originResource, and resourceGrant
// adding a subject requires calculating all matching CRP's for that one subject (filtered for its classes)
// adding a CRP requires reconciling the new CRP for all subjects
// adding an RG requires looking up all applicable originID's via the index and allowing their references
type grantGraph map[subject]map[targetResourceGroup]map[CRPID]map[originResourceID]map[targetNamespacedName]map[purpose]grant

type namespace string
type originResourceGroup schema.GroupResource

// originResourceIndex is used to find applicable grants to add/remove a ReferenceGrant ID for the grantors list
type originResourceIndex map[namespace]map[originResourceGroup]originResourceID

// in-memory GraphStore
type AuthStore struct {
	graph               grantGraph
	originResourceIndex originResourceIndex
	mutex               sync.RWMutex
}

func NewGraphStore() *AuthStore {
	return &AuthStore{graph: make(grantGraph)}
}

func (s *AuthStore) CheckAuthz(sar authorizationv1.SubjectAccessReview) (bool, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	user := sar.Spec.User
	groups := sar.Spec.Groups
	trg := targetResourceGroup{
		Resource: sar.Spec.ResourceAttributes.Resource,
		Group:    sar.Spec.ResourceAttributes.Group,
	}
	orig := originResourceID(strings.Join(sar.Spec.Extra[ExtraOriginKey], ", ")) // TODO: where can we get this value from the client request properly?
	nn := targetNamespacedName{
		namespace: sar.GetName(),
		name:      sar.GetNamespace(),
	}
	p := purpose(strings.Join(sar.Spec.Extra[ExtraPurposeKey], ", ")) // TODO: where can we get this value from the client request properly?
	v := verb(sar.Spec.ResourceAttributes.Verb)

	var hasErr error
	for _, g := range groups {
		allowed, err := s.lookup(subject(g), trg, orig, nn, p, v)
		if err != nil {
			hasErr = err // TODO: multi error
		}
		if allowed {
			return true, hasErr
		}
	}
	allowed, err := s.lookup(subject(user), trg, orig, nn, p, v)
	if err != nil {
		hasErr = err // TODO: multi error
	}
	return allowed, hasErr
}

// Graph Lookup is perfomed as follows:
// 1. Start by attempting to find the subject (subj) in the authorization graph. 
//    If it's not found, it returns false indicating that the subject doesn't have any grants.
// 
// 2. It then looks for the target resource group (trg) within the subject's map in the graph.
//    If not found, it returns false.
//
// 3. Next, it iterates through all the CRPIDs (ClusterReferencePatternIDs) within the target resource group's map.
//    For each CRPID, it checks if an origin resource ID is provided. 
//    If not, it iterates through all possible origin resource IDs and checks for grants.
//
// 4. If an origin resource ID is provided, it directly looks up grants for that specific origin resource ID.
// 5. If a matching grant is found for the given purpose and verb, it returns true.
// 6. If no matching grants are found, it continues the iteration. If it exhausts all possibilities, it returns false.
func (s *AuthStore) lookup(subj subject, trg targetResourceGroup, orig originResourceID, nn targetNamespacedName, p purpose, v verb) (bool, error) {
	// if verbs, ok := s.graph[targetResource]; ok {
	// 	if allowed, ok := verbs[verb]; ok {
	// 		return allowed, nil
	// 	}
	// }

	trgMap, ok := s.graph[subj]
	if !ok {
		return false, nil
	}
	crpMap, ok := trgMap[trg]
	if !ok {
		return false, nil
	}
	// lookup from every possible CRP for this targetResourceGroup
	for crp, origMap := range crpMap {
		if orig == "" {
			// no originResourceID provided, fallback and check for any possibilities
			for o, nnMap := range origMap {
				if pgMap, ok := nnMap[nn]; ok {
					if checkGrant(pgMap, p, v) {
						return true, nil
					}
				}
			}
		} else {
			if nnMap, ok := origMap[orig]; ok {
				if pgMap, ok := nnMap[nn]; ok {
					if checkGrant(pgMap, p, v) {
						return true, nil
					}
				}
			}
		}
	}

	return false, nil
}

func checkGrant(pgMap map[purpose]grant, p purpose, v verb) bool {
	if g, ok := pgMap[p]; ok {
		if len(g.grantors) > 0 {
			for _, gv := range g.verbs {
				if v == gv {
					return true
				}
			}
		}
	}
	return false
}

func (s *AuthStore) UpsertGrant(subj subject, trg targetResourceGroup, crpid CRPID, orig originResourceID, nn targetNamespacedName, p purpose, v []verb, grantor string) {
	// if _, ok := s.graph[resource]; !ok {
	// 	s.graph[resource] = make(map[string]bool)
	// }
	// s.graph[resource][verb] = allowed
}

func (s *AuthStore) RevokeClusterReferenceConsumer(crc *v1alpha1.ClusterReferenceConsumer) {

}

func (s *AuthStore) RevokeClusterReferencePattern(crp *v1alpha1.ClusterReferenceGrant) {

}

func (s *AuthStore) RevokeOriginObject(orig originResourceID) {

}

func (s *AuthStore) RevokeReferenceGrant(rg *v1alpha1.ReferenceGrant) {

}
