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

package handlers

import (
	"encoding/json"
	"net/http"

	authorizationv1 "k8s.io/api/authorization/v1"
	"sigs.k8s.io/referencegrant-poc/pkg/store"
)

func AuthzHandler(store *store.AuthStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var sar authorizationv1.SubjectAccessReview
		err := decoder.Decode(&sar)
		if err != nil {
			http.Error(w, "Error decoding SubjectAccessReview", http.StatusBadRequest)
			return
		}
		// I acquire the lock in CheckAuthz but maybe we should do this here instead?
		// ... Call store.CheckAuthz(sar)
		// ... Construct AuthorizationResponse
	}
}

func HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
