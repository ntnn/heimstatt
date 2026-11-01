module heimstatt.dev

go 1.27.0

tool codeberg.org/ntnn/mindl

replace heimstatt.dev/apis => ./apis

require (
	github.com/kcp-dev/kcp-operator/sdk v0.10.2
	github.com/stretchr/testify v1.12.1
	heimstatt.dev/apis v0.0.0-00010101000000-000000000000
	k8s.io/api v0.37.2
	k8s.io/apimachinery v0.37.2
	k8s.io/utils v0.0.0-20261023123148-ca04ba58ecab
	sigs.k8s.io/gateway-api v1.7.0
)

require (
	codeberg.org/ntnn/mindl v0.1.4 // indirect
	github.com/fxamacker/cbor/v2 v2.9.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kcp-dev/logicalcluster/v3 v3.0.5 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/net v0.61.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/client-go v0.37.1 // indirect
	k8s.io/klog/v2 v2.140.0 // indirect
	k8s.io/kube-openapi v0.0.0-20260721132016-d427ff9ee9ad // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.4.2 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)
