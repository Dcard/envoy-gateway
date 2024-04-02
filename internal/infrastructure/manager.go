// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package infrastructure

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	clicfg "sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes"
	"github.com/envoyproxy/gateway/internal/ir"
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
		clientConf := clicfg.GetConfigOrDie()

		httpClient, err := newHTTPClient(clientConf)
		if err != nil {
			return nil, err
		}

		cli, err := client.New(clientConf, client.Options{
			Scheme:     envoygateway.GetScheme(),
			HTTPClient: httpClient,
		})
		if err != nil {
			return nil, err
		}

		mgr = kubernetes.NewInfra(cli, cfg)
	} else {
		// Kube is the only supported provider type for now.
		return nil, fmt.Errorf("unsupported provider type %v", cfg.EnvoyGateway.Provider.Type)
	}

	return mgr, nil
}

func newHTTPClient(config *rest.Config) (*http.Client, error) {
	transport, err := rest.TransportFor(config)
	if err != nil {
		return nil, fmt.Errorf("could not create HTTP transport: %w", err)
	}

	transport = rewriteAPIVersion{transport}

	return &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}, nil
}

const (
	gatewayAPIPath        = "/apis/gateway.networking.k8s.io/"
	gatewayAPIPathV1      = gatewayAPIPath + "v1/"
	gatewayAPIPathV1Beta1 = gatewayAPIPath + "v1beta1/"
)

type rewriteAPIVersion struct {
	http.RoundTripper
}

func (r rewriteAPIVersion) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.HasPrefix(req.URL.Path, gatewayAPIPathV1) {
		req.URL.Path = strings.Replace(req.URL.Path, gatewayAPIPathV1, gatewayAPIPathV1Beta1, 1)
	}

	return r.RoundTripper.RoundTrip(req)
}
