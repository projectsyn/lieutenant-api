module github.com/projectsyn/lieutenant

go 1.13

require (
	github.com/AlekSi/pointer v1.1.0
	github.com/cosmtrek/air v1.12.0
	github.com/deepmap/oapi-codegen v1.3.4
	github.com/getkin/kin-openapi v0.2.0
	github.com/labstack/echo/v4 v4.1.11
	github.com/mattn/go-isatty v0.0.11 // indirect
	github.com/stretchr/testify v1.4.0
	golang.org/x/sys v0.0.0-20200120151820-655fe14d7479 // indirect
)

// replace github.com/deepmap/oapi-codegen => ../../GitHub/deepmap/oapi-codegen/
replace github.com/deepmap/oapi-codegen => github.com/vshn/oapi-codegen v1.3.5-0.20200120165139-75a50d5f3093
