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
package status

import (
	"testing"

	contour_api_v1 "github.com/projectcontour/contour/apis/projectcontour/v1"
	contour_api_v1alpha1 "github.com/projectcontour/contour/apis/projectcontour/v1alpha1"
	"github.com/projectcontour/contour/pkg/fixture"
	"github.com/projectcontour/contour/pkg/k8s"
	"github.com/stretchr/testify/assert"
	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	gatewayapi_v1alpha1 "sigs.k8s.io/gateway-api/apis/v1alpha1"
)

type testCacheEntry struct {
	ConditionCache

	ID string
}

func (t testCacheEntry) AsStatusUpdate() k8s.StatusUpdate {
	return k8s.StatusUpdate{
		NamespacedName: types.NamespacedName{
			Namespace: core_v1.NamespaceDefault,
			Name:      t.ID,
		},
	}
}

var _ CacheEntry = &testCacheEntry{}

func TestCacheAcquisition(t *testing.T) {
	ext := &contour_api_v1alpha1.ExtensionService{
		ObjectMeta: fixture.ObjectMeta("test/ext"),
	}
	proxy := &contour_api_v1.HTTPProxy{
		ObjectMeta: fixture.ObjectMeta("test/proxy"),
	}
	httpRoute := &gatewayapi_v1alpha1.HTTPRoute{
		ObjectMeta: fixture.ObjectMeta("test/httproute"),
	}
	cache := NewCache(types.NamespacedName{Name: "contour", Namespace: "projectcontour"})

	// Initial acquisition should be nil.
	assert.Nil(t, cache.Get(proxy))
	assert.Nil(t, cache.Get(httpRoute))
	assert.Nil(t, cache.Get(ext))

	newEntry := testCacheEntry{ID: "AA483012-A14F-4644-A3C9-FDBAAFA958C0"}
	cache.Put(proxy, &newEntry)
	cache.Put(ext, &newEntry)
	cache.Put(httpRoute, &newEntry)

	cachedEntry := cache.Get(proxy)
	assert.Equal(t, &newEntry, cachedEntry)

	cachedEntry = cache.Get(httpRoute)
	assert.Equal(t, &newEntry, cachedEntry)

	updates := cache.GetStatusUpdates()
	assert.Equal(t, 3, len(updates))
	assert.Equal(t, newEntry.ID, updates[0].NamespacedName.Name)

	assert.Equal(t, 3, len(cache.entries))
	assert.Equal(t, 1, len(cache.entries["HTTPProxy"]))
	assert.Equal(t, 1, len(cache.entries["ExtensionService"]))
	assert.Equal(t, 1, len(cache.entries["HTTPRoute"]))
}
