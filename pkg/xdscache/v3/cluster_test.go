// Copyright Project Contour Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v3

import (
	"testing"
	"time"

	envoy_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_extensions_upstream_http_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/duration"
	contour_api_v1 "github.com/projectcontour/contour/apis/projectcontour/v1"
	envoy_v3 "github.com/projectcontour/contour/pkg/envoy/v3"
	"github.com/projectcontour/contour/pkg/protobuf"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestClusterCacheContents(t *testing.T) {
	tests := map[string]struct {
		contents map[string]*envoy_cluster_v3.Cluster
		want     []proto.Message
	}{
		"empty": {
			contents: nil,
			want:     nil,
		},
		"simple": {
			contents: clustermap(
				&envoy_cluster_v3.Cluster{
					Name:                 "default/kuard/443/da39a3ee5e",
					AltStatName:          "default_kuard_443",
					ClusterDiscoveryType: envoy_v3.ClusterDiscoveryType(envoy_cluster_v3.Cluster_EDS),
					EdsClusterConfig: &envoy_cluster_v3.Cluster_EdsClusterConfig{
						EdsConfig:   envoy_v3.ConfigSource("contour"),
						ServiceName: "default/kuard",
					},
				}),
			want: []proto.Message{
				cluster(&envoy_cluster_v3.Cluster{
					Name:                 "default/kuard/443/da39a3ee5e",
					AltStatName:          "default_kuard_443",
					ClusterDiscoveryType: envoy_v3.ClusterDiscoveryType(envoy_cluster_v3.Cluster_EDS),
					EdsClusterConfig: &envoy_cluster_v3.Cluster_EdsClusterConfig{
						EdsConfig:   envoy_v3.ConfigSource("contour"),
						ServiceName: "default/kuard",
					},
				}),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var cc ClusterCache
			cc.Update(tc.contents)
			got := cc.Contents()
			protobuf.ExpectEqual(t, tc.want, got)
		})
	}
}

func TestClusterCacheQuery(t *testing.T) {
	tests := map[string]struct {
		contents map[string]*envoy_cluster_v3.Cluster
		query    []string
		want     []proto.Message
	}{
		"exact match": {
			contents: clustermap(
				&envoy_cluster_v3.Cluster{
					Name:                 "default/kuard/443/da39a3ee5e",
					AltStatName:          "default_kuard_443",
					ClusterDiscoveryType: envoy_v3.ClusterDiscoveryType(envoy_cluster_v3.Cluster_EDS),
					EdsClusterConfig: &envoy_cluster_v3.Cluster_EdsClusterConfig{
						EdsConfig:   envoy_v3.ConfigSource("contour"),
						ServiceName: "default/kuard",
					},
				}),
			query: []string{"default/kuard/443/da39a3ee5e"},
			want: []proto.Message{
				cluster(&envoy_cluster_v3.Cluster{
					Name:                 "default/kuard/443/da39a3ee5e",
					AltStatName:          "default_kuard_443",
					ClusterDiscoveryType: envoy_v3.ClusterDiscoveryType(envoy_cluster_v3.Cluster_EDS),
					EdsClusterConfig: &envoy_cluster_v3.Cluster_EdsClusterConfig{
						EdsConfig:   envoy_v3.ConfigSource("contour"),
						ServiceName: "default/kuard",
					},
				}),
			},
		},
		"partial match": {
			contents: clustermap(
				&envoy_cluster_v3.Cluster{
					Name:                 "default/kuard/443/da39a3ee5e",
					AltStatName:          "default_kuard_443",
					ClusterDiscoveryType: envoy_v3.ClusterDiscoveryType(envoy_cluster_v3.Cluster_EDS),
					EdsClusterConfig: &envoy_cluster_v3.Cluster_EdsClusterConfig{
						EdsConfig:   envoy_v3.ConfigSource("contour"),
						ServiceName: "default/kuard",
					},
				}),
			query: []string{"default/kuard/443/da39a3ee5e", "foo/bar/baz"},
			want: []proto.Message{
				cluster(&envoy_cluster_v3.Cluster{
					Name:                 "default/kuard/443/da39a3ee5e",
					AltStatName:          "default_kuard_443",
					ClusterDiscoveryType: envoy_v3.ClusterDiscoveryType(envoy_cluster_v3.Cluster_EDS),
					EdsClusterConfig: &envoy_cluster_v3.Cluster_EdsClusterConfig{
						EdsConfig:   envoy_v3.ConfigSource("contour"),
						ServiceName: "default/kuard",
					},
				}),
			},
		},
		"no match": {
			contents: clustermap(
				&envoy_cluster_v3.Cluster{
					Name:                 "default/kuard/443/da39a3ee5e",
					AltStatName:          "default_kuard_443",
					ClusterDiscoveryType: envoy_v3.ClusterDiscoveryType(envoy_cluster_v3.Cluster_EDS),
					EdsClusterConfig: &envoy_cluster_v3.Cluster_EdsClusterConfig{
						EdsConfig:   envoy_v3.ConfigSource("contour"),
						ServiceName: "default/kuard",
					},
				}),
			query: []string{"foo/bar/baz"},
			want:  nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var cc ClusterCache
			cc.Update(tc.contents)
			got := cc.Query(tc.query)
			protobuf.ExpectEqual(t, tc.want, got)
		})
	}
}

func TestClusterVisit(t *testing.T) {
	tests := map[string]struct {
		objs []interface{}
		want map[string]*envoy_cluster_v3.Cluster
	}{
		"nothing": {
			objs: nil,
			want: map[string]*envoy_cluster_v3.Cluster{},
		},
		"single unnamed service": {
			objs: []interface{}{
				&v1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kuard",
						Namespace: "default",
					},
					Spec: v1beta1.IngressSpec{
						Backend: backend("kuard", 443),
					},
				},
				service("default", "kuard",
					v1.ServicePort{
						Protocol:   "TCP",
						Port:       443,
						TargetPort: intstr.FromInt(8443),
					},
				),
			},
			want: clustermap(
				&envoy_cluster_v3.Cluster{
					Name:                 "default/kuard/443/da39a3ee5e",
					AltStatName:          "default_kuard_443",
					ClusterDiscoveryType: envoy_v3.ClusterDiscoveryType(envoy_cluster_v3.Cluster_EDS),
					EdsClusterConfig: &envoy_cluster_v3.Cluster_EdsClusterConfig{
						EdsConfig:   envoy_v3.ConfigSource("contour"),
						ServiceName: "default/kuard",
					},
				}),
		},
		"single named service": {
			objs: []interface{}{
				&v1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kuard",
						Namespace: "default",
					},
					Spec: v1beta1.IngressSpec{
						Backend: &v1beta1.IngressBackend{
							ServiceName: "kuard",
							ServicePort: intstr.FromString("https"),
						},
					},
				},
				service("default", "kuard",
					v1.ServicePort{
						Name:       "https",
						Protocol:   "TCP",
						Port:       443,
						TargetPort: intstr.FromInt(8443),
					},
				),
			},
			want: clustermap(
				&envoy_cluster_v3.Cluster{
					Name:                 "default/kuard/443/da39a3ee5e",
					AltStatName:          "default_kuard_443",
					ClusterDiscoveryType: envoy_v3.ClusterDiscoveryType(envoy_cluster_v3.Cluster_EDS),
					EdsClusterConfig: &envoy_cluster_v3.Cluster_EdsClusterConfig{
						EdsConfig:   envoy_v3.ConfigSource("contour"),
						ServiceName: "default/kuard/https",
					},
				}),
		},
		"h2c upstream": {
			objs: []interface{}{
				&v1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kuard",
						Namespace: "default",
					},
					Spec: v1beta1.IngressSpec{
						Backend: &v1beta1.IngressBackend{
							ServiceName: "kuard",
							ServicePort: intstr.FromString("http"),
						},
					},
				},
				serviceWithAnnotations(
					"default",
					"kuard",
					map[string]string{
						"projectcontour.io/upstream-protocol.h2c": "80,http",
					},
					v1.ServicePort{
						Protocol: "TCP",
						Name:     "http",
						Port:     80,
					},
				),
			},
			want: clustermap(
				&envoy_cluster_v3.Cluster{
					Name:                 "default/kuard/80/da39a3ee5e",
					AltStatName:          "default_kuard_80",
					ClusterDiscoveryType: envoy_v3.ClusterDiscoveryType(envoy_cluster_v3.Cluster_EDS),
					EdsClusterConfig: &envoy_cluster_v3.Cluster_EdsClusterConfig{
						EdsConfig:   envoy_v3.ConfigSource("contour"),
						ServiceName: "default/kuard/http",
					},
					TypedExtensionProtocolOptions: map[string]*any.Any{
						"envoy.extensions.upstreams.http.v3.HttpProtocolOptions": protobuf.MustMarshalAny(
							&envoy_extensions_upstream_http_v3.HttpProtocolOptions{
								UpstreamProtocolOptions: &envoy_extensions_upstream_http_v3.HttpProtocolOptions_ExplicitHttpConfig_{
									ExplicitHttpConfig: &envoy_extensions_upstream_http_v3.HttpProtocolOptions_ExplicitHttpConfig{
										ProtocolConfig: &envoy_extensions_upstream_http_v3.HttpProtocolOptions_ExplicitHttpConfig_Http2ProtocolOptions{},
									},
								},
							}),
					},
				},
			),
		},
		"long namespace and service name": {
			objs: []interface{}{
				&v1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "webserver-1-unimatrix-zero-one",
						Namespace: "beurocratic-company-test-domain-1",
					},
					Spec: v1beta1.IngressSpec{
						Backend: backend("tiny-cog-department-test-instance", 443),
					},
				},
				service("beurocratic-company-test-domain-1", "tiny-cog-department-test-instance",
					v1.ServicePort{
						Name:       "svc-0",
						Protocol:   "TCP",
						Port:       443,
						TargetPort: intstr.FromInt(8443),
					},
				),
			},
			want: clustermap(
				&envoy_cluster_v3.Cluster{
					Name:                 "beurocra-7fe4b4/tiny-cog-7fe4b4/443/da39a3ee5e",
					AltStatName:          "beurocratic-company-test-domain-1_tiny-cog-department-test-instance_443",
					ClusterDiscoveryType: envoy_v3.ClusterDiscoveryType(envoy_cluster_v3.Cluster_EDS),
					EdsClusterConfig: &envoy_cluster_v3.Cluster_EdsClusterConfig{
						EdsConfig:   envoy_v3.ConfigSource("contour"),
						ServiceName: "beurocratic-company-test-domain-1/tiny-cog-department-test-instance/svc-0",
					},
				}),
		},
		"two service ports": {
			objs: []interface{}{
				&contour_api_v1.HTTPProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple",
						Namespace: "default",
					},
					Spec: contour_api_v1.HTTPProxySpec{
						VirtualHost: &contour_api_v1.VirtualHost{
							Fqdn: "www.example.com",
						},
						Routes: []contour_api_v1.Route{{
							Conditions: []contour_api_v1.MatchCondition{{
								Prefix: "/",
							}},
							Services: []contour_api_v1.Service{{
								Name: "backend",
								Port: 80,
							}, {
								Name: "backend",
								Port: 8080,
							}},
						}},
					},
				},
				service("default", "backend", v1.ServicePort{
					Name:       "http",
					Protocol:   "TCP",
					Port:       80,
					TargetPort: intstr.FromInt(6502),
				}, v1.ServicePort{
					Name:       "alt",
					Protocol:   "TCP",
					Port:       8080,
					TargetPort: intstr.FromString("9001"),
				}),
			},
			want: clustermap(
				&envoy_cluster_v3.Cluster{
					Name:                 "default/backend/80/da39a3ee5e",
					AltStatName:          "default_backend_80",
					ClusterDiscoveryType: envoy_v3.ClusterDiscoveryType(envoy_cluster_v3.Cluster_EDS),
					EdsClusterConfig: &envoy_cluster_v3.Cluster_EdsClusterConfig{
						EdsConfig:   envoy_v3.ConfigSource("contour"),
						ServiceName: "default/backend/http",
					},
				},
				&envoy_cluster_v3.Cluster{
					Name:                 "default/backend/8080/da39a3ee5e",
					AltStatName:          "default_backend_8080",
					ClusterDiscoveryType: envoy_v3.ClusterDiscoveryType(envoy_cluster_v3.Cluster_EDS),
					EdsClusterConfig: &envoy_cluster_v3.Cluster_EdsClusterConfig{
						EdsConfig:   envoy_v3.ConfigSource("contour"),
						ServiceName: "default/backend/alt",
					},
				},
			),
		},
		"httpproxy with simple path healthcheck": {
			objs: []interface{}{
				&contour_api_v1.HTTPProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple",
						Namespace: "default",
					},
					Spec: contour_api_v1.HTTPProxySpec{
						VirtualHost: &contour_api_v1.VirtualHost{
							Fqdn: "www.example.com",
						},
						Routes: []contour_api_v1.Route{{
							HealthCheckPolicy: &contour_api_v1.HTTPHealthCheckPolicy{
								Path: "/healthy",
							},
							Conditions: []contour_api_v1.MatchCondition{{
								Prefix: "/",
							}},
							Services: []contour_api_v1.Service{{
								Name: "backend",
								Port: 80,
							}},
						}},
					},
				},
				service("default", "backend", v1.ServicePort{
					Name:       "http",
					Protocol:   "TCP",
					Port:       80,
					TargetPort: intstr.FromInt(6502),
				}),
			},
			want: clustermap(
				&envoy_cluster_v3.Cluster{
					Name:                 "default/backend/80/c184349821",
					AltStatName:          "default_backend_80",
					ClusterDiscoveryType: envoy_v3.ClusterDiscoveryType(envoy_cluster_v3.Cluster_EDS),
					EdsClusterConfig: &envoy_cluster_v3.Cluster_EdsClusterConfig{
						EdsConfig:   envoy_v3.ConfigSource("contour"),
						ServiceName: "default/backend/http",
					},
					HealthChecks: []*envoy_core_v3.HealthCheck{{
						Timeout:            &duration.Duration{Seconds: 2},
						Interval:           &duration.Duration{Seconds: 10},
						UnhealthyThreshold: protobuf.UInt32(3),
						HealthyThreshold:   protobuf.UInt32(2),
						HealthChecker: &envoy_core_v3.HealthCheck_HttpHealthCheck_{
							HttpHealthCheck: &envoy_core_v3.HealthCheck_HttpHealthCheck{
								Path: "/healthy",
								Host: "contour-envoy-healthcheck",
							},
						},
					}},
					IgnoreHealthOnHostRemoval: true,
				},
			),
		},
		"httpproxy with custom healthcheck": {
			objs: []interface{}{
				&contour_api_v1.HTTPProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple",
						Namespace: "default",
					},
					Spec: contour_api_v1.HTTPProxySpec{
						VirtualHost: &contour_api_v1.VirtualHost{
							Fqdn: "www.example.com",
						},
						Routes: []contour_api_v1.Route{{
							HealthCheckPolicy: &contour_api_v1.HTTPHealthCheckPolicy{
								Host:                    "foo-bar-host",
								Path:                    "/healthy",
								TimeoutSeconds:          99,
								IntervalSeconds:         98,
								UnhealthyThresholdCount: 97,
								HealthyThresholdCount:   96,
							},
							Conditions: []contour_api_v1.MatchCondition{{
								Prefix: "/",
							}},
							Services: []contour_api_v1.Service{{
								Name: "backend",
								Port: 80,
							}},
						}},
					},
				},
				service("default", "backend", v1.ServicePort{
					Name:       "http",
					Protocol:   "TCP",
					Port:       80,
					TargetPort: intstr.FromInt(6502),
				}),
			},
			want: clustermap(
				&envoy_cluster_v3.Cluster{
					Name:                 "default/backend/80/7f8051653a",
					AltStatName:          "default_backend_80",
					ClusterDiscoveryType: envoy_v3.ClusterDiscoveryType(envoy_cluster_v3.Cluster_EDS),
					EdsClusterConfig: &envoy_cluster_v3.Cluster_EdsClusterConfig{
						EdsConfig:   envoy_v3.ConfigSource("contour"),
						ServiceName: "default/backend/http",
					},
					HealthChecks: []*envoy_core_v3.HealthCheck{{
						Timeout:            &duration.Duration{Seconds: 99},
						Interval:           &duration.Duration{Seconds: 98},
						UnhealthyThreshold: protobuf.UInt32(97),
						HealthyThreshold:   protobuf.UInt32(96),
						HealthChecker: &envoy_core_v3.HealthCheck_HttpHealthCheck_{
							HttpHealthCheck: &envoy_core_v3.HealthCheck_HttpHealthCheck{
								Path: "/healthy",
								Host: "foo-bar-host",
							},
						},
					}},
					IgnoreHealthOnHostRemoval: true,
				},
			),
		},
		"httpproxy with RoundRobin lb algorithm": {
			objs: []interface{}{
				&contour_api_v1.HTTPProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple",
						Namespace: "default",
					},
					Spec: contour_api_v1.HTTPProxySpec{
						VirtualHost: &contour_api_v1.VirtualHost{
							Fqdn: "www.example.com",
						},
						Routes: []contour_api_v1.Route{{
							LoadBalancerPolicy: &contour_api_v1.LoadBalancerPolicy{
								Strategy: "RoundRobin",
							},
							Conditions: []contour_api_v1.MatchCondition{{
								Prefix: "/",
							}},
							Services: []contour_api_v1.Service{{
								Name: "backend",
								Port: 80,
							}},
						}},
					},
				},
				service("default", "backend", v1.ServicePort{
					Name:       "http",
					Protocol:   "TCP",
					Port:       80,
					TargetPort: intstr.FromInt(6502),
				}),
			},
			want: clustermap(
				&envoy_cluster_v3.Cluster{
					Name:                 "default/backend/80/da39a3ee5e",
					AltStatName:          "default_backend_80",
					ClusterDiscoveryType: envoy_v3.ClusterDiscoveryType(envoy_cluster_v3.Cluster_EDS),
					EdsClusterConfig: &envoy_cluster_v3.Cluster_EdsClusterConfig{
						EdsConfig:   envoy_v3.ConfigSource("contour"),
						ServiceName: "default/backend/http",
					},
				},
			),
		},
		"httpproxy with WeightedLeastRequest lb algorithm": {
			objs: []interface{}{
				&contour_api_v1.HTTPProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple",
						Namespace: "default",
					},
					Spec: contour_api_v1.HTTPProxySpec{
						VirtualHost: &contour_api_v1.VirtualHost{
							Fqdn: "www.example.com",
						},
						Routes: []contour_api_v1.Route{{
							LoadBalancerPolicy: &contour_api_v1.LoadBalancerPolicy{
								Strategy: "WeightedLeastRequest",
							},
							Conditions: []contour_api_v1.MatchCondition{{
								Prefix: "/",
							}},
							Services: []contour_api_v1.Service{{
								Name: "backend",
								Port: 80,
							}},
						}},
					},
				},
				service("default", "backend", v1.ServicePort{
					Name:       "http",
					Protocol:   "TCP",
					Port:       80,
					TargetPort: intstr.FromInt(6502),
				}),
			},
			want: clustermap(
				&envoy_cluster_v3.Cluster{
					Name:                 "default/backend/80/8bf87fefba",
					AltStatName:          "default_backend_80",
					ClusterDiscoveryType: envoy_v3.ClusterDiscoveryType(envoy_cluster_v3.Cluster_EDS),
					EdsClusterConfig: &envoy_cluster_v3.Cluster_EdsClusterConfig{
						EdsConfig:   envoy_v3.ConfigSource("contour"),
						ServiceName: "default/backend/http",
					},
					LbPolicy: envoy_cluster_v3.Cluster_LEAST_REQUEST,
				},
			),
		},
		"httpproxy with Random lb algorithm": {
			objs: []interface{}{
				&contour_api_v1.HTTPProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple",
						Namespace: "default",
					},
					Spec: contour_api_v1.HTTPProxySpec{
						VirtualHost: &contour_api_v1.VirtualHost{
							Fqdn: "www.example.com",
						},
						Routes: []contour_api_v1.Route{{
							LoadBalancerPolicy: &contour_api_v1.LoadBalancerPolicy{
								Strategy: "Random",
							},
							Conditions: []contour_api_v1.MatchCondition{{
								Prefix: "/",
							}},
							Services: []contour_api_v1.Service{{
								Name: "backend",
								Port: 80,
							}},
						}},
					},
				},
				service("default", "backend", v1.ServicePort{
					Name:       "http",
					Protocol:   "TCP",
					Port:       80,
					TargetPort: intstr.FromInt(6502),
				}),
			},
			want: clustermap(
				&envoy_cluster_v3.Cluster{
					Name:                 "default/backend/80/58d888c08a",
					AltStatName:          "default_backend_80",
					ClusterDiscoveryType: envoy_v3.ClusterDiscoveryType(envoy_cluster_v3.Cluster_EDS),
					EdsClusterConfig: &envoy_cluster_v3.Cluster_EdsClusterConfig{
						EdsConfig:   envoy_v3.ConfigSource("contour"),
						ServiceName: "default/backend/http",
					},
					LbPolicy: envoy_cluster_v3.Cluster_RANDOM,
				},
			),
		},
		"httpproxy with RequestHash lb algorithm and valid header hash option": {
			objs: []interface{}{
				&contour_api_v1.HTTPProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple",
						Namespace: "default",
					},
					Spec: contour_api_v1.HTTPProxySpec{
						VirtualHost: &contour_api_v1.VirtualHost{
							Fqdn: "www.example.com",
						},
						Routes: []contour_api_v1.Route{{
							LoadBalancerPolicy: &contour_api_v1.LoadBalancerPolicy{
								Strategy: "RequestHash",
								RequestHashPolicies: []contour_api_v1.RequestHashPolicy{
									{
										HeaderHashOptions: &contour_api_v1.HeaderHashOptions{
											HeaderName: "X-Custom-Header",
										},
									},
								},
							},
							Conditions: []contour_api_v1.MatchCondition{{
								Prefix: "/",
							}},
							Services: []contour_api_v1.Service{{
								Name: "backend",
								Port: 80,
							}},
						}},
					},
				},
				service("default", "backend", v1.ServicePort{
					Name:       "http",
					Protocol:   "TCP",
					Port:       80,
					TargetPort: intstr.FromInt(6502),
				}),
			},
			want: clustermap(
				&envoy_cluster_v3.Cluster{
					Name:                 "default/backend/80/1a2ffc1fef",
					AltStatName:          "default_backend_80",
					ClusterDiscoveryType: envoy_v3.ClusterDiscoveryType(envoy_cluster_v3.Cluster_EDS),
					EdsClusterConfig: &envoy_cluster_v3.Cluster_EdsClusterConfig{
						EdsConfig:   envoy_v3.ConfigSource("contour"),
						ServiceName: "default/backend/http",
					},
					LbPolicy: envoy_cluster_v3.Cluster_RING_HASH,
				},
			),
		},
		// Removed testcase - "httpproxy with differing lb algorithms"
		// HTTPProxy has LB algorithm as a route-level construct, so it's not possible.
		"httpproxy with unknown lb algorithm": {
			objs: []interface{}{
				&contour_api_v1.HTTPProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple",
						Namespace: "default",
					},
					Spec: contour_api_v1.HTTPProxySpec{
						VirtualHost: &contour_api_v1.VirtualHost{
							Fqdn: "www.example.com",
						},
						Routes: []contour_api_v1.Route{{
							LoadBalancerPolicy: &contour_api_v1.LoadBalancerPolicy{
								Strategy: "lulz",
							},
							Conditions: []contour_api_v1.MatchCondition{{
								Prefix: "/",
							}},
							Services: []contour_api_v1.Service{{
								Name: "backend",
								Port: 80,
							}},
						}},
					},
				},
				service("default", "backend", v1.ServicePort{
					Name:       "http",
					Protocol:   "TCP",
					Port:       80,
					TargetPort: intstr.FromInt(6502),
				}),
			},
			want: clustermap(
				&envoy_cluster_v3.Cluster{
					Name:                 "default/backend/80/da39a3ee5e",
					AltStatName:          "default_backend_80",
					ClusterDiscoveryType: envoy_v3.ClusterDiscoveryType(envoy_cluster_v3.Cluster_EDS),
					EdsClusterConfig: &envoy_cluster_v3.Cluster_EdsClusterConfig{
						EdsConfig:   envoy_v3.ConfigSource("contour"),
						ServiceName: "default/backend/http",
					},
				},
			),
		},
		"circuitbreaker annotations": {
			objs: []interface{}{
				&v1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kuard",
						Namespace: "default",
					},
					Spec: v1beta1.IngressSpec{
						Backend: &v1beta1.IngressBackend{
							ServiceName: "kuard",
							ServicePort: intstr.FromString("http"),
						},
					},
				},
				serviceWithAnnotations(
					"default",
					"kuard",
					map[string]string{
						"projectcontour.io/max-connections":      "9000",
						"projectcontour.io/max-pending-requests": "4096",
						"projectcontour.io/max-requests":         "404",
						"projectcontour.io/max-retries":          "7",
					},
					v1.ServicePort{
						Protocol: "TCP",
						Name:     "http",
						Port:     80,
					},
				),
			},
			want: clustermap(
				&envoy_cluster_v3.Cluster{
					Name:                 "default/kuard/80/da39a3ee5e",
					AltStatName:          "default_kuard_80",
					ClusterDiscoveryType: envoy_v3.ClusterDiscoveryType(envoy_cluster_v3.Cluster_EDS),
					EdsClusterConfig: &envoy_cluster_v3.Cluster_EdsClusterConfig{
						EdsConfig:   envoy_v3.ConfigSource("contour"),
						ServiceName: "default/kuard/http",
					},
					CircuitBreakers: &envoy_cluster_v3.CircuitBreakers{
						Thresholds: []*envoy_cluster_v3.CircuitBreakers_Thresholds{{
							MaxConnections:     protobuf.UInt32(9000),
							MaxPendingRequests: protobuf.UInt32(4096),
							MaxRequests:        protobuf.UInt32(404),
							MaxRetries:         protobuf.UInt32(7),
						}},
					},
				},
			),
		},
		"projectcontour.io/num-retries annotation": {
			objs: []interface{}{
				&v1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kuard",
						Namespace: "default",
						Annotations: map[string]string{
							"projectcontour.io/num-retries": "7",
							"projectcontour.io/retry-on":    "gateway-error",
						},
					},
					Spec: v1beta1.IngressSpec{
						Backend: &v1beta1.IngressBackend{
							ServiceName: "kuard",
							ServicePort: intstr.FromString("https"),
						},
					},
				},
				service("default", "kuard",
					v1.ServicePort{
						Name:       "https",
						Protocol:   "TCP",
						Port:       443,
						TargetPort: intstr.FromInt(8443),
					},
				),
			},
			want: clustermap(
				&envoy_cluster_v3.Cluster{
					Name:                 "default/kuard/443/da39a3ee5e",
					AltStatName:          "default_kuard_443",
					ClusterDiscoveryType: envoy_v3.ClusterDiscoveryType(envoy_cluster_v3.Cluster_EDS),
					EdsClusterConfig: &envoy_cluster_v3.Cluster_EdsClusterConfig{
						EdsConfig:   envoy_v3.ConfigSource("contour"),
						ServiceName: "default/kuard/https",
					},
				}),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			root := buildDAG(t, tc.objs...)
			got := visitClusters(root)
			protobuf.ExpectEqual(t, tc.want, got)
		})
	}
}

func service(ns, name string, ports ...v1.ServicePort) *v1.Service {
	return serviceWithAnnotations(ns, name, nil, ports...)
}

func serviceWithAnnotations(ns, name string, annotations map[string]string, ports ...v1.ServicePort) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   ns,
			Annotations: annotations,
		},
		Spec: v1.ServiceSpec{
			Ports: ports,
		},
	}
}

func cluster(c *envoy_cluster_v3.Cluster) *envoy_cluster_v3.Cluster {
	// NOTE: Keep this in sync with envoy.defaultCluster().
	defaults := &envoy_cluster_v3.Cluster{
		ConnectTimeout: protobuf.Duration(250 * time.Millisecond),
		CommonLbConfig: envoy_v3.ClusterCommonLBConfig(),
		LbPolicy:       envoy_cluster_v3.Cluster_ROUND_ROBIN,
	}

	proto.Merge(defaults, c)
	return defaults
}

func clustermap(clusters ...*envoy_cluster_v3.Cluster) map[string]*envoy_cluster_v3.Cluster {
	m := make(map[string]*envoy_cluster_v3.Cluster)
	for _, c := range clusters {
		m[c.Name] = cluster(c)
	}
	return m
}
