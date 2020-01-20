module github.com/projectsyn/lieutenant

go 1.13

require (
	github.com/AlekSi/pointer v1.1.0
	github.com/deepmap/oapi-codegen v1.3.4
	github.com/getkin/kin-openapi v0.2.0
	github.com/labstack/echo/v4 v4.1.11
)

// replace github.com/deepmap/oapi-codegen => ../../GitHub/deepmap/oapi-codegen/
replace github.com/deepmap/oapi-codegen => github.com/vshn/oapi-codegen v1.3.5-0.20200120090250-02fcfbfe6f01
