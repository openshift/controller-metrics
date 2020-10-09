module github.com/openshift/controller-metrics

go 1.13

require (
	github.com/prometheus/client_golang v1.7.1
	github.com/stretchr/testify v1.6.1
	k8s.io/client-go v0.18.4
	sigs.k8s.io/controller-runtime v0.3.0
)

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190918200256-06eb1244587a
