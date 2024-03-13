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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	authorizationv1 "k8s.io/api/authorization/v1"
	"sigs.k8s.io/referencegrant-poc/pkg/store"
)

func sendNotAuthorizedResponse(w http.ResponseWriter, reason string) {
	resp := authorizationv1.SubjectAccessReview{
		Status: authorizationv1.SubjectAccessReviewStatus{
			Allowed: false,
			Reason:  reason,
		},
	}
	responseBytes, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Write(responseBytes)

}

func AuthzHandler(store *store.AuthStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		content, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error in Read of request body : %s", err)
			sendNotAuthorizedResponse(w, "Error in Read of request body")
		}
		rawContent := json.RawMessage(string(content))

		sar := authorizationv1.SubjectAccessReview{}
		err = json.Unmarshal(content, &sar)
		if err != nil {
			log.Printf("Failed to unmarshal: %v", err)
			sendNotAuthorizedResponse(w, err.Error())
			return
		}

		if sar.Spec.ResourceAttributes == nil {
			sendNotAuthorizedResponse(w, "Not authorizing nonResourceAttributes")
			return
		}
		var print bool
		if strings.Contains(sar.Spec.User, "kong-controller") && sar.Spec.ResourceAttributes.Resource == "secrets" && strings.HasPrefix(sar.Spec.ResourceAttributes.Name, "demo") {
			print = true
			log.Printf("Request body : %s\n", rawContent)
		}

		allowed, _ := store.CheckAuthz(sar)
		sarResponseStatus := authorizationv1.SubjectAccessReviewStatus{
			Allowed: allowed,
		}
		if !allowed {
			sarResponseStatus.Reason = fmt.Sprintf("Referential authorizer did not allow Subject \"%s\" to %s %s/%s/%s/%s ", sar.Spec.User, sar.Spec.ResourceAttributes.Verb, sar.Spec.ResourceAttributes.Group, sar.Spec.ResourceAttributes.Resource, sar.Spec.ResourceAttributes.Namespace, sar.Spec.ResourceAttributes.Name)
		}
		sar.Status = sarResponseStatus

		sar.Spec = authorizationv1.SubjectAccessReviewSpec{}
		responseBytes, _ := json.Marshal(sar)
		if print {
			log.Printf("Response body : %s\n", responseBytes)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(responseBytes)
	}
}

func HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Server is healthy"))
}
