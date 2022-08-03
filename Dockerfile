# syntax = docker/dockerfile:experimental

FROM docker.io/golang:1.19 as build
ARG VERSION

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=$GOPATH/pkg/mod go mod download

COPY . .

RUN --mount=type=cache,target=$HOME/.cache/go-build make test
RUN --mount=type=cache,target=$HOME/.cache/go-build make build

FROM gcr.io/distroless/static:nonroot

COPY --from=build /app/lieutenant-api /usr/local/bin/

ENTRYPOINT [ "/usr/local/bin/lieutenant-api" ]
