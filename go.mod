module github.com/che-incubator/configbump

go 1.12

require (
	github.com/alexflint/go-arg v1.3.0
	github.com/operator-framework/operator-sdk v0.19.0
	k8s.io/api v0.18.4
	k8s.io/apimachinery v0.18.4
	sigs.k8s.io/controller-runtime v0.6.1
)

replace (
    k8s.io/client-go => k8s.io/client-go v0.18.4
)
