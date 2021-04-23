module github.com/projectcontour/contour

go 1.15

require (
	github.com/ahmetb/gen-crd-api-reference-docs v0.3.0
	github.com/envoyproxy/go-control-plane v0.9.9-0.20210111201334-f1f47757da33
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.2
	github.com/google/uuid v1.1.2
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/prometheus/client_golang v1.9.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.15.0
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	google.golang.org/genproto v0.0.0-20201110150050-8816d57aaa9a
	google.golang.org/grpc v1.27.1
	google.golang.org/protobuf v1.25.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v0.21.0
	k8s.io/klog/v2 v2.8.0
	k8s.io/utils v0.0.0-20210305010621-2afb4311ab10
	sigs.k8s.io/controller-runtime v0.8.2
	sigs.k8s.io/controller-tools v0.4.1
	sigs.k8s.io/gateway-api v0.2.0
	sigs.k8s.io/kustomize/kyaml v0.1.1
)

replace github.com/projectcontour/contour => github.com/touchuyht/contour v1.14.3
