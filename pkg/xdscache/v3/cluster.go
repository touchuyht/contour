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
	"sort"
	"sync"

	envoy_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/golang/protobuf/proto"
	"github.com/projectcontour/contour/pkg/contour"
	"github.com/projectcontour/contour/pkg/dag"
	"github.com/projectcontour/contour/pkg/envoy"
	envoy_v3 "github.com/projectcontour/contour/pkg/envoy/v3"
	"github.com/projectcontour/contour/pkg/protobuf"
	"github.com/projectcontour/contour/pkg/sorter"
)

// ClusterCache manages the contents of the gRPC CDS cache.
type ClusterCache struct {
	mu     sync.Mutex
	values map[string]*envoy_cluster_v3.Cluster
	contour.Cond
}

// Update replaces the contents of the cache with the supplied map.
func (c *ClusterCache) Update(v map[string]*envoy_cluster_v3.Cluster) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.values = v
	c.Cond.Notify()
}

// Contents returns a copy of the cache's contents.
func (c *ClusterCache) Contents() []proto.Message {
	c.mu.Lock()
	defer c.mu.Unlock()
	var values []*envoy_cluster_v3.Cluster
	for _, v := range c.values {
		values = append(values, v)
	}
	sort.Stable(sorter.For(values))
	return protobuf.AsMessages(values)
}

func (c *ClusterCache) Query(names []string) []proto.Message {
	c.mu.Lock()
	defer c.mu.Unlock()
	var values []*envoy_cluster_v3.Cluster
	for _, n := range names {
		// if the cluster is not registered we cannot return
		// a blank cluster because each cluster has a required
		// discovery type; DNS, EDS, etc. We cannot determine the
		// correct value for this property from the cluster's name
		// provided by the query so we must not return a blank cluster.
		if v, ok := c.values[n]; ok {
			values = append(values, v)
		}
	}
	sort.Stable(sorter.For(values))
	return protobuf.AsMessages(values)
}

func (*ClusterCache) TypeURL() string { return resource.ClusterType }

func (c *ClusterCache) OnChange(root *dag.DAG) {
	clusters := visitClusters(root)
	c.Update(clusters)
}

type clusterVisitor struct {
	clusters map[string]*envoy_cluster_v3.Cluster
}

// visitCluster produces a map of *envoy_cluster_v3.Clusters.
func visitClusters(root dag.Vertex) map[string]*envoy_cluster_v3.Cluster {
	cv := clusterVisitor{
		clusters: make(map[string]*envoy_cluster_v3.Cluster),
	}
	cv.visit(root)
	return cv.clusters
}

func (v *clusterVisitor) visit(vertex dag.Vertex) {
	switch cluster := vertex.(type) {
	case *dag.Cluster:
		name := envoy.Clustername(cluster)
		if _, ok := v.clusters[name]; !ok {
			v.clusters[name] = envoy_v3.Cluster(cluster)
		}
	case *dag.ExtensionCluster:
		name := cluster.Name
		if _, ok := v.clusters[name]; !ok {
			v.clusters[name] = envoy_v3.ExtensionCluster(cluster)
		}
	}

	// recurse into children of v
	vertex.Visit(v.visit)
}
