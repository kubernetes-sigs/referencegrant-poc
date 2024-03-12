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
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

type GatewayEventsHandler struct {
	c      *Controller
	logger logr.Logger
}

func NewGatewayEventsHandler(c *Controller) *GatewayEventsHandler {
	return &GatewayEventsHandler{
		c:      c,
		logger: c.log.WithName("eventHandlers").WithName("gateway"),
	}
}

func (h *GatewayEventsHandler) Create(ctx context.Context, e event.CreateEvent, q workqueue.RateLimitingInterface) {
	h.constructEventFromGateway(e.Object, q)
}

func (h *GatewayEventsHandler) Update(ctx context.Context, e event.UpdateEvent, q workqueue.RateLimitingInterface) {
	h.constructEventFromGateway(e.ObjectNew, q)
}

func (h *GatewayEventsHandler) Delete(ctx context.Context, e event.DeleteEvent, q workqueue.RateLimitingInterface) {
	h.constructEventFromGateway(e.Object, q)
}

func (h *GatewayEventsHandler) Generic(ctx context.Context, e event.GenericEvent, q workqueue.RateLimitingInterface) {
	h.constructEventFromGateway(e.Object, q)
}

func (h *GatewayEventsHandler) constructEventFromGateway(obj client.Object, q workqueue.RateLimitingInterface) {
	gw := obj.(*gatewayv1.Gateway)
	name := fmt.Sprintf("Gateway/%s", gw.Name)
	fromKey := "gateway.networking.k8s.io/gateways"
	// ctx := context.TODO()
	// crgList := &v1a1.ClusterReferenceGrantList{}
	// err := h.c.crClient.List(ctx, crgList)
	// if err != nil {
	// 	h.c.log.Error(err, "could not list ClusterReferenceGrants")
	// 	return
	// }
	// keys := getApplicableKeys(crgList, fromKey)
	// for key := range keys {
	// 	q.AddRateLimited(reconcile.Request{NamespacedName: types.NamespacedName{Name: name, Namespace: key}})
	// }
	q.AddRateLimited(reconcile.Request{NamespacedName: types.NamespacedName{Name: name, Namespace: fromKey}})

}
