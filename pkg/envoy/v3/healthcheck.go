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
	"time"

	envoy_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/projectcontour/contour/pkg/dag"
	"github.com/projectcontour/contour/pkg/envoy"
	"github.com/projectcontour/contour/pkg/protobuf"
)

// httpHealthCheck returns a *envoy_core_v3.HealthCheck value for HTTP Routes
func httpHealthCheck(cluster *dag.Cluster) *envoy_core_v3.HealthCheck {
	hc := cluster.HTTPHealthCheckPolicy
	host := envoy.HCHost
	if hc.Host != "" {
		host = hc.Host
	}

	// TODO(dfc) why do we need to specify our own default, what is the default
	// that envoy applies if these fields are left nil?
	return &envoy_core_v3.HealthCheck{
		Timeout:            durationOrDefault(hc.Timeout, envoy.HCTimeout),
		Interval:           durationOrDefault(hc.Interval, envoy.HCInterval),
		UnhealthyThreshold: protobuf.UInt32OrDefault(hc.UnhealthyThreshold, envoy.HCUnhealthyThreshold),
		HealthyThreshold:   protobuf.UInt32OrDefault(hc.HealthyThreshold, envoy.HCHealthyThreshold),
		HealthChecker: &envoy_core_v3.HealthCheck_HttpHealthCheck_{
			HttpHealthCheck: &envoy_core_v3.HealthCheck_HttpHealthCheck{
				Path: hc.Path,
				Host: host,
			},
		},
	}
}

// tcpHealthCheck returns a *envoy_core_v3.HealthCheck value for TCPProxies
func tcpHealthCheck(cluster *dag.Cluster) *envoy_core_v3.HealthCheck {
	hc := cluster.TCPHealthCheckPolicy

	return &envoy_core_v3.HealthCheck{
		Timeout:            durationOrDefault(hc.Timeout, envoy.HCTimeout),
		Interval:           durationOrDefault(hc.Interval, envoy.HCInterval),
		UnhealthyThreshold: protobuf.UInt32OrDefault(hc.UnhealthyThreshold, envoy.HCUnhealthyThreshold),
		HealthyThreshold:   protobuf.UInt32OrDefault(hc.HealthyThreshold, envoy.HCHealthyThreshold),
		HealthChecker: &envoy_core_v3.HealthCheck_TcpHealthCheck_{
			TcpHealthCheck: &envoy_core_v3.HealthCheck_TcpHealthCheck{},
		},
	}
}

func durationOrDefault(d, def time.Duration) *duration.Duration {
	if d != 0 {
		return protobuf.Duration(d)
	}
	return protobuf.Duration(def)
}
