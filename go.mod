module github.com/projectsyn/lieutenant-api

go 1.13

require (
	github.com/AlekSi/pointer v1.1.0
	github.com/cosmtrek/air v1.26.0
	github.com/deepmap/oapi-codegen v1.4.1
	github.com/getkin/kin-openapi v0.35.0
	github.com/labstack/echo/v4 v4.2.2
	github.com/projectsyn/lieutenant-operator v0.5.2
	github.com/rakyll/statik v0.1.7
	github.com/stretchr/testify v1.7.0
	github.com/taion809/haikunator v0.0.0-20150324135039-4e414e676fd1
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	sigs.k8s.io/controller-runtime v0.8.3
)

replace k8s.io/client-go => k8s.io/client-go v0.20.4

replace github.com/docker/docker => github.com/moby/moby v1.13.1 // Required by github.com/operator-framework/operator-lifecycle-manager, from lieutenant-operator

// replace github.com/projectsyn/lieutenant-operator => github.com/projectsyn/lieutenant-operator v0.0.5
