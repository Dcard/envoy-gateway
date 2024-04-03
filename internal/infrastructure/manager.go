// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package infrastructure

import (
	"context"
	"fmt"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes"
	"github.com/envoyproxy/gateway/internal/ir"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
	clicfg "sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

var _ Manager = (*kubernetes.Infra)(nil)

// Manager provides the scaffolding for managing infrastructure.
type Manager interface {
	// CreateOrUpdateProxyInfra creates or updates infra.
	CreateOrUpdateProxyInfra(ctx context.Context, infra *ir.Infra) error
	// DeleteProxyInfra deletes infra.
	DeleteProxyInfra(ctx context.Context, infra *ir.Infra) error
	// CreateOrUpdateRateLimitInfra creates or updates rate limit infra.
	CreateOrUpdateRateLimitInfra(ctx context.Context) error
	// DeleteRateLimitInfra deletes rate limit infra.
	DeleteRateLimitInfra(ctx context.Context) error
}

// NewManager returns a new infrastructure Manager.
func NewManager(cfg *config.Server) (Manager, error) {
	var mgr Manager
	if cfg.EnvoyGateway.Provider.Type == v1alpha1.ProviderTypeKubernetes {
		cli, err := client.New(clicfg.GetConfigOrDie(), client.Options{
			Scheme: envoygateway.GetScheme(),
		})
		if err != nil {
			return nil, err
		}

		var watchClient client.WithWatch = nopWatchClient{Client: cli}
		watchClient = interceptor.NewClient(watchClient, interceptor.Funcs{
			Get: func(ctx context.Context, client client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
				b1Obj := replaceGatewayV1Object(obj)

				if err := client.Get(ctx, key, b1Obj, opts...); err != nil {
					return err
				}

				copyGatewayObjectProps(obj, b1Obj)

				return nil
			},
			List: func(ctx context.Context, client client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
				b1Obj := replaceGatewayV1ObjectList(list)

				if err := client.List(ctx, b1Obj, opts...); err != nil {
					return err
				}

				copyGatewayObjectListProps(list, b1Obj)

				return nil
			},
			Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
				return client.Create(ctx, replaceGatewayV1Object(obj), opts...)
			},
			Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				return client.Update(ctx, replaceGatewayV1Object(obj), opts...)
			},
			Delete: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
				return client.Delete(ctx, replaceGatewayV1Object(obj), opts...)
			},
			DeleteAllOf: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.DeleteAllOfOption) error {
				return client.DeleteAllOf(ctx, replaceGatewayV1Object(obj), opts...)
			},
			Patch: func(ctx context.Context, client client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
				return client.Patch(ctx, replaceGatewayV1Object(obj), patch, opts...)
			},
		})

		mgr = kubernetes.NewInfra(watchClient, cfg)
	} else {
		// Kube is the only supported provider type for now.
		return nil, fmt.Errorf("unsupported provider type %v", cfg.EnvoyGateway.Provider.Type)
	}

	return mgr, nil
}

var _ client.WithWatch = nopWatchClient{}

type nopWatchClient struct {
	client.Client
}

func (n nopWatchClient) Watch(context.Context, client.ObjectList, ...client.ListOption) (watch.Interface, error) {
	return nil, fmt.Errorf("watch is not supported")
}

var gwv1b1APIVersion = gwv1b1.GroupVersion.String()

func gwv1b1TypeMeta(kind string) metav1.TypeMeta {
	return metav1.TypeMeta{APIVersion: gwv1b1APIVersion, Kind: kind}
}

func replaceGatewayV1Object(obj client.Object) client.Object {
	switch obj := obj.(type) {
	case *gwv1.Gateway:
		return &gwv1b1.Gateway{
			TypeMeta:   gwv1b1TypeMeta(gatewayapi.KindGateway),
			ObjectMeta: obj.ObjectMeta,
			Spec:       obj.Spec,
			Status:     obj.Status,
		}
	case *gwv1.HTTPRoute:
		return &gwv1b1.HTTPRoute{
			TypeMeta:   gwv1b1TypeMeta(gatewayapi.KindHTTPRoute),
			ObjectMeta: obj.ObjectMeta,
			Spec:       obj.Spec,
			Status:     obj.Status,
		}
	case *gwv1.GatewayClass:
		return &gwv1b1.GatewayClass{
			TypeMeta:   gwv1b1TypeMeta(gatewayapi.KindGatewayClass),
			ObjectMeta: obj.ObjectMeta,
			Spec:       obj.Spec,
			Status:     obj.Status,
		}
	default:
		return obj
	}
}

func copyGatewayObjectProps(dst, src client.Object) {
	switch dst := dst.(type) {
	case *gwv1.Gateway:
		src := src.(*gwv1b1.Gateway)
		dst.ObjectMeta = src.ObjectMeta
		dst.Spec = src.Spec
		dst.Status = src.Status
	case *gwv1.HTTPRoute:
		src := src.(*gwv1b1.HTTPRoute)
		dst.ObjectMeta = src.ObjectMeta
		dst.Spec = src.Spec
		dst.Status = src.Status
	case *gwv1.GatewayClass:
		src := src.(*gwv1b1.GatewayClass)
		dst.ObjectMeta = src.ObjectMeta
		dst.Spec = src.Spec
		dst.Status = src.Status
	}
}

func replaceGatewayV1ObjectList(list client.ObjectList) client.ObjectList {
	switch list := list.(type) {
	case *gwv1.GatewayList:
		items := make([]gwv1b1.Gateway, len(list.Items))

		for i, item := range list.Items {
			items[i] = gwv1b1.Gateway{
				TypeMeta:   gwv1b1TypeMeta(gatewayapi.KindGateway),
				ObjectMeta: item.ObjectMeta,
				Spec:       list.Items[i].Spec,
				Status:     list.Items[i].Status,
			}
		}

		return &gwv1b1.GatewayList{
			TypeMeta: gwv1b1TypeMeta("GatewayList"),
			ListMeta: list.ListMeta,
			Items:    items,
		}

	case *gwv1.HTTPRouteList:
		items := make([]gwv1b1.HTTPRoute, len(list.Items))

		for i, item := range list.Items {
			items[i] = gwv1b1.HTTPRoute{
				TypeMeta:   gwv1b1TypeMeta(gatewayapi.KindGateway),
				ObjectMeta: item.ObjectMeta,
				Spec:       list.Items[i].Spec,
				Status:     list.Items[i].Status,
			}
		}

		return &gwv1b1.HTTPRouteList{
			TypeMeta: gwv1b1TypeMeta("HTTPRouteList"),
			ListMeta: list.ListMeta,
			Items:    items,
		}
	case *gwv1.GatewayClassList:
		items := make([]gwv1b1.GatewayClass, len(list.Items))

		for i, item := range list.Items {
			items[i] = gwv1b1.GatewayClass{
				TypeMeta:   gwv1b1TypeMeta(gatewayapi.KindGatewayClass),
				ObjectMeta: item.ObjectMeta,
				Spec:       list.Items[i].Spec,
				Status:     list.Items[i].Status,
			}
		}

		return &gwv1b1.GatewayClassList{
			TypeMeta: gwv1b1TypeMeta("GatewayClassList"),
			ListMeta: list.ListMeta,
			Items:    items,
		}

	default:
		return list
	}
}

func copyGatewayObjectListProps(dst, src client.ObjectList) {
	switch dst := dst.(type) {
	case *gwv1.GatewayList:
		src := src.(*gwv1b1.GatewayList)
		dst.ListMeta = src.ListMeta
		dst.Items = make([]gwv1.Gateway, len(src.Items))

		for i, item := range src.Items {
			item := item
			copyGatewayObjectProps(&dst.Items[i], &item)
		}

	case *gwv1.HTTPRouteList:
		src := src.(*gwv1b1.HTTPRouteList)
		dst.ListMeta = src.ListMeta
		dst.Items = make([]gwv1.HTTPRoute, len(src.Items))

		for i, item := range src.Items {
			item := item
			copyGatewayObjectProps(&dst.Items[i], &item)
		}

	case *gwv1.GatewayClassList:
		src := src.(*gwv1b1.GatewayClassList)
		dst.ListMeta = src.ListMeta
		dst.Items = make([]gwv1.GatewayClass, len(src.Items))

		for i, item := range src.Items {
			item := item
			copyGatewayObjectProps(&dst.Items[i], &item)
		}
	}
}
