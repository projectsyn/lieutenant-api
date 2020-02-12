# Project Syn: Lieutenant API

Rest API to provide services like inventory, cluster registry, tenant management and
GitOps helper.

**Please note that this project is in it's early stages and under active development**.

## OpenAPI Spec

The API is specified in [OpenAPI 3](https://swagger.io/docs/specification/about/) format.
It's available in the file [openapi.yaml](openapi.yaml) in the root folder of this project.

## Run API locally

To run the API on your local workstation, follow these steps:

```
export KUBECONFIG=~/.kube/myconfig
export NAMESPACE=syn-lieutenant
make run
```

The `kubeconfig` must grant access to the cluster.

Check with `curl localhost:8080/healthz` if the API is responding.

## Example queries

Done with [HTTPie](https://httpie.org/), but also works with plain `curl`.

_Create Tenant_
```
http localhost:8080/tenants Authorization:"Bearer $(kubectl get secrets test-token-zzzzz -o json | jq ".data.token" -r | base64 --decode)" displayName="Syn Corp"
```

_Query Tenants_

```
http localhost:8080/tenants Authorization:"Bearer $(kubectl get secrets test-token-zzzzz -o json | jq ".data.token" -r | base64 --decode)"
```

* `test-token-zzzzz` is the token of a ServiceAccount with all the needed RBAC
  rights on the underlying Kubernetes cluster

## Development

### API Versioning

There is no API versioning exposed. Internally the "API Evolution" approach is
used which is described in this excellent blog post:
[API Versioning Has No "Right Way"](https://apisyouwonthate.com/blog/api-versioning-has-no-right-way).

### API Mocking

To test the API before writing actual code, API mocking can be used. One cool tool
to do that is [Prism](https://github.com/stoplightio/prism). Example:

```
docker run --init --rm -p 4010:4010 -v $(pwd):/tmp stoplight/prism:3 mock -h 0.0.0.0 "/tmp/openapi.yaml"

curl http://localhost:4010/tenants
```

### Code Generation

As the API spec is written with OpenAPI, the [OpenAPI Generator](https://openapi-generator.tech/) is used to generate the Go boilerplate code.

### Object Name Generation

Some API endpoints store data in Kubernetes objects. These objects must be
named like this:

`pwgen -A -B 6 1`

### Links

Some good links which help with OpenAPI development:

* [OpenAPI.Tools](https://openapi.tools/)
* [OpenAPI (Swagger) Editor](https://marketplace.visualstudio.com/items?itemName=42Crunch.vscode-openapi) (Visual Studio Code Extension)
* [OpenAPI Preview](https://marketplace.visualstudio.com/items?itemName=zoellner.openapi-preview) (Visual Studio Code Extension)
