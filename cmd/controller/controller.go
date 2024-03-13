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

package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/jsonpath"
	"k8s.io/klog/v2/klogr"
	"k8s.io/klog/v2/textlogger"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	v1a1 "sigs.k8s.io/referencegrant-poc/apis/v1alpha1"
	"sigs.k8s.io/referencegrant-poc/pkg/store"
)

const (
	labelKeyPatternName = "reference.authorization.k8s.io/pattern-name"
)

type Controller struct {
	dClient  *dynamic.DynamicClient
	crClient client.Client
	log      logr.Logger
	store    *store.AuthStore
}

func NewController(authStore *store.AuthStore) *Controller {
	lConfig := textlogger.NewConfig()

	c := &Controller{
		log:   textlogger.NewLogger(lConfig),
		store: authStore,
	}
	ctrl.SetLogger(klogr.New())

	c.log.Info("Initializing Controller")

	kConfig := ctrl.GetConfigOrDie()
	scheme := scheme.Scheme
	v1a1.AddToScheme(scheme)
	gatewayv1.AddToScheme(scheme)

	dClient, err := dynamic.NewForConfig(kConfig)
	if err != nil {
		c.log.Error(err, "could not create Dynamic client")
		os.Exit(1)
	}

	c.dClient = dClient

	manager, err := ctrl.NewManager(kConfig, ctrl.Options{Scheme: scheme})
	if err != nil {
		c.log.Error(err, "could not create manager")
		os.Exit(1)
	}

	c.crClient = manager.GetClient()

	// TODO: Add selective ClusterRole and RoleBinding watchers here
	err = ctrl.NewControllerManagedBy(manager).
		Named("referencegrant-poc").
		Watches(&v1a1.ClusterReferenceConsumer{}, NewClusterReferenceConsumerHandler(c)).
		Watches(&v1a1.ClusterReferenceGrant{}, NewClusterReferenceGrantHandler(c)).
		Watches(&v1a1.ReferenceGrant{}, NewReferenceGrantHandler(c)).
		Watches(&gatewayv1.Gateway{}, NewGatewayEventsHandler(c)).
		Complete(c)

	if err != nil {
		c.log.Error(err, "could not setup controller")
		os.Exit(1)
	}

	if err := manager.Start(ctrl.SetupSignalHandler()); err != nil {
		c.log.Error(err, "could not start manager")
		os.Exit(1)
	}

	return c
}

func getApplicableKeys(crgList *v1a1.ClusterReferenceGrantList, from string) sets.Set[string] {
	keys := make(sets.Set[string])
	for _, crg := range crgList.Items {
		origin := fmt.Sprintf("%s/%s", crg.From.Group, crg.From.Resource)
		// Early exit if origin is not the same
		if from != origin {
			continue
		}
		if len(crg.Versions) == 0 {
			continue
		}
		for _, ref := range crg.Versions[0].References {
			target := fmt.Sprintf("%s/%s", ref.To.Group, ref.To.Resource)
			key := fmt.Sprintf("%s;%s;%s", origin, target, ref.For)
			keys.Insert(key)
		}
	}
	return keys
}
func (c *Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	c.log.Info("Reconciling for", "name", req.NamespacedName.Name)

	crcList := &v1a1.ClusterReferenceConsumerList{}
	err := c.crClient.List(ctx, crcList)
	if err != nil {
		c.log.Error(err, "could not list ClusterReferenceConsumers")
		return ctrl.Result{}, err
	}
	crgList := &v1a1.ClusterReferenceGrantList{}
	err = c.crClient.List(ctx, crgList)
	if err != nil {
		c.log.Error(err, "could not list ClusterReferenceGrants")
		return ctrl.Result{}, err
	}

	rgList := &v1a1.ReferenceGrantList{}
	err = c.crClient.List(ctx, rgList)
	if err != nil {
		c.log.Error(err, "could not list ReferenceGrants")
		return ctrl.Result{}, err
	}

	// Currently when we enter the reconcile with Gateway event, we only have the "From".
	// Hence we recalculate all applicable keys in graph with "gateway.networking.k8s.io.gateway" as "From"
	c.log.Info(fmt.Sprintf("req: %s", req.NamespacedName.Namespace))
	keys := make(sets.Set[string])
	if strings.Split(req.NamespacedName.Name, "/")[0] == "Gateway" {
		keys = getApplicableKeys(crgList, req.NamespacedName.Namespace)
	} else {
		keys.Insert(req.NamespacedName.Namespace)
	}
	c.log.Info(fmt.Sprintf("keys: %v", keys))

	for fromToForKey := range keys {
		c.log.Info("Reconciling for", "name", fromToForKey)
		crcSubjects := []v1a1.Subject{}
		for _, crc := range crcList.Items {
			for _, ref := range crc.References {
				// qKey := NewQueueKey(ref.From.Group, ref.From.Resource, ref.To.Group, ref.To.Resource, ref.For)
				origin := fmt.Sprintf("%s/%s", ref.From.Group, ref.From.Resource)
				target := fmt.Sprintf("%s/%s", ref.To.Group, ref.To.Resource)
				key := fmt.Sprintf("%s;%s;%s", origin, target, ref.For)
				if key == fromToForKey {
					if crc.Subject.Kind == "ServiceAccount" {
						crc.Subject.Kind = "User"
						crc.Subject.Name = fmt.Sprintf("system:serviceaccount:%s:%s", crc.Subject.Namespace, crc.Subject.Name)
						crc.Subject.Namespace = ""
					}
					crcSubjects = append(crcSubjects, crc.Subject)
					break
				}
			}
		}
		// map between FromNamespace to a map of ToNamespace to []ResourceName for this particular fromToFor key
		crossNamespaceGrants := map[string]map[string]sets.Set[string]{}
		for _, rg := range rgList.Items {
			origin := fmt.Sprintf("%s/%s", rg.From.Group, rg.From.Resource)
			target := fmt.Sprintf("%s/%s", rg.To.Group, rg.To.Resource)
			key := fmt.Sprintf("%s;%s;%s", origin, target, rg.For)
			if key == fromToForKey {
				if _, ok := crossNamespaceGrants[rg.From.Namespace]; !ok {
					crossNamespaceGrants[rg.From.Namespace] = make(map[string]sets.Set[string])
				}
				if _, ok := crossNamespaceGrants[rg.From.Namespace][rg.Namespace]; !ok {
					crossNamespaceGrants[rg.From.Namespace][rg.Namespace] = make(sets.Set[string])
				}
				crossNamespaceGrants[rg.From.Namespace][rg.Namespace].Insert(rg.To.Names...)
			}
		}

		referencePaths := []*v1a1.ReferencePath{}
		var fromVersion string
		for _, crg := range crgList.Items {
			origin := fmt.Sprintf("%s/%s", crg.From.Group, crg.From.Resource)
			// Early exit if origin is not the same
			if strings.Split(fromToForKey, ";")[0] != origin {
				continue
			}
			if len(crg.Versions) == 0 {
				c.log.Info("Skipping clusterReferenceGrant with no versions", "name", crg.Name)
				continue
			}
			fromVersion = crg.Versions[0].Version
			// TODO: Handle versions, currently taking only the first versions in the list
			for _, ref := range crg.Versions[0].References {
				target := fmt.Sprintf("%s/%s", ref.To.Group, ref.To.Resource)
				key := fmt.Sprintf("%s;%s;%s", origin, target, ref.For)
				if key == fromToForKey {
					referencePaths = append(referencePaths, &ref)
				}
			}
		}

		// At this point, all references declaration in ReferencePaths are relevant for us.
		// Follow Reference Paths to recalculate the graph.
		origin := strings.Split(strings.Split(fromToForKey, ";")[0], "/")
		fromGroup, fromResource := origin[0], origin[1]

		var referencedResources []reference
		for _, refPath := range referencePaths {
			// TODO: Have informers for each target resource of a ClusterReferenceGrant
			targetGVR := schema.GroupVersionResource{Group: fromGroup, Version: fromVersion, Resource: fromResource}
			targetList, err := c.dClient.Resource(targetGVR).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				c.log.Error(err, "failed to list target for ClusterReferenceGrant", targetGVR)
				return ctrl.Result{}, err
			}
			refResources, err := c.getReferences(targetList, refPath.Path)
			if err != nil {
				c.log.Error(err, "failed to follow references for path", refPath.Path, targetGVR)
				return ctrl.Result{}, err
			}
			referencedResources = append(referencedResources, refResources...)
		}

		// For the first POC, we want to recalculate the from-to-for key e2e. Clearing the key first
		c.store.ClearGraphKey(fromToForKey)

		finalRefResources := []reference{}
		for _, refResource := range referencedResources {
			if refResource.FromNamespace != refResource.ToNamespace {
				if allowedToNamespaces, ok := crossNamespaceGrants[refResource.FromNamespace]; ok {
					if allowedNamesForNamespace, exist := allowedToNamespaces[refResource.ToNamespace]; exist {
						if allowedNamesForNamespace.Has(refResource.Name) {
							finalRefResources = append(finalRefResources, refResource)
						}
					}
				}
			} else {
				finalRefResources = append(finalRefResources, refResource)
			}
		}
		for _, refResource := range finalRefResources {
			c.store.UpsertGrant(fromToForKey, types.NamespacedName{Namespace: refResource.ToNamespace, Name: refResource.Name}, crcSubjects)
		}
		c.log.V(0).Info("Reconciliation finished", "GraphKey", fromToForKey)
		c.log.V(0).Info(fmt.Sprintf("Graph is:\n%v\n", c.store.GetGraph()))
		c.log.V(0).Info(fmt.Sprintf("SubjectIndex is: %v\n", c.store.GetSubjectIndex()))
	}

	return ctrl.Result{}, nil
}

func generateQueueKey(fromGroup, fromResource, toGroup, toResource, forReason string) types.NamespacedName {
	nn := types.NamespacedName{Name: forReason}
	nn.Namespace = fmt.Sprintf("%s/%s-%s/%s")
	return nn
}

type reference struct {
	Group         string
	Resource      string
	FromNamespace string
	ToNamespace   string
	Name          string
}

func (c *Controller) getReferences(list *unstructured.UnstructuredList, path string) ([]reference, error) {
	refs := []reference{}
	for _, item := range list.Items {
		j := jsonpath.New("test").AllowMissingKeys(true)
		err := j.Parse(fmt.Sprintf("{%s}", path))
		if err != nil {
			c.log.Error(err, "error parsing JSON Path")
			return refs, err
		}
		results := new(bytes.Buffer)
		err = j.Execute(results, item.UnstructuredContent())
		if err != nil {
			c.log.Error(err, "error finding results with JSON Path")
			return refs, err
		}

		rawRefs := strings.Split(results.String(), " ")

		for _, rr := range rawRefs {
			jr := map[string]string{}
			if len(rr) == 0 {
				continue
			}
			err = json.Unmarshal([]byte(rr), &jr)
			if err != nil {
				c.log.Error(err, "error finding results with JSON Path")
			}
			// The part below is commented in favour of the decision to
			// requiring the ClusterRefGrant to specify the group and resource and
			// limit the json path to only pull names that match that group and resource or kind.
			// This is not feasible using the current jsonPath implementation so we are likely to use CEL for this.

			// group, hasGroup := jr["group"]
			// if !hasGroup {
			// 	c.log.Info("Missing group in reference", "ref", jr)
			// 	continue
			// }
			// resource, hasResource := jr["resource"]
			// if !hasResource {
			// 	kind, hasKind := jr["kind"]
			// 	if !hasKind {
			// 		c.log.Info("Missing kind or resource in reference", "ref", jr)
			// 		continue
			// 	}
			// 	gvr, _ := meta.UnsafeGuessKindToResource(schema.GroupVersionKind{Group: group, Version: "v1", Kind: kind})
			// 	resource = gvr.Resource
			// }

			namespace, hasNamespace := jr["namespace"]
			if !hasNamespace {
				namespace = item.GetNamespace()
			}

			name, hasName := jr["name"]
			if !hasName {
				c.log.Info("Missing name in reference", "ref", jr)
				continue
			}
			refs = append(refs, reference{
				FromNamespace: item.GetNamespace(),
				ToNamespace:   namespace,
				Name:          name,
			})
		}
	}

	return refs, nil
}
