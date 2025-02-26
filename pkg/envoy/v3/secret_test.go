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

	envoy_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_tls_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/projectcontour/contour/pkg/dag"
	"github.com/projectcontour/contour/pkg/envoy"
	"github.com/projectcontour/contour/pkg/protobuf"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSecret(t *testing.T) {
	tests := map[string]struct {
		secret *dag.Secret
		want   *envoy_tls_v3.Secret
	}{
		"simple secret": {
			secret: &dag.Secret{
				Object: &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple",
						Namespace: "default",
					},
					Data: map[string][]byte{
						v1.TLSCertKey:       []byte("cert"),
						v1.TLSPrivateKeyKey: []byte("key"),
					},
				},
			},
			want: &envoy_tls_v3.Secret{
				Name: "default/simple/cd1b506996",
				Type: &envoy_tls_v3.Secret_TlsCertificate{
					TlsCertificate: &envoy_tls_v3.TlsCertificate{
						PrivateKey: &envoy_core_v3.DataSource{
							Specifier: &envoy_core_v3.DataSource_InlineBytes{
								InlineBytes: []byte("key"),
							},
						},
						CertificateChain: &envoy_core_v3.DataSource{
							Specifier: &envoy_core_v3.DataSource_InlineBytes{
								InlineBytes: []byte("cert"),
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := Secret(tc.secret)
			protobuf.ExpectEqual(t, tc.want, got)
		})
	}
}

func TestSecretname(t *testing.T) {
	tests := map[string]struct {
		secret *dag.Secret
		want   string
	}{
		"simple": {
			secret: &dag.Secret{
				Object: &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple",
						Namespace: "default",
					},
					Data: map[string][]byte{
						v1.TLSCertKey:       []byte("cert"),
						v1.TLSPrivateKeyKey: []byte("key"),
					},
				},
			},
			want: "default/simple/cd1b506996",
		},
		"far too long": {
			secret: &dag.Secret{
				Object: &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "must-be-in-want-of-a-wife",
						Namespace: "it-is-a-truth-universally-acknowledged-that-a-single-man-in-possession-of-a-good-fortune",
					},
					Data: map[string][]byte{
						v1.TLSCertKey:       []byte("cert"),
						v1.TLSPrivateKeyKey: []byte("key"),
					},
				},
			},
			want: "it-is-a-truth-7e164b/must-be-in-wa-7e164b/cd1b506996",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := envoy.Secretname(tc.secret)
			assert.Equal(t, tc.want, got)
		})
	}
}
