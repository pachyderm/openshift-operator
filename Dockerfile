# Build the manager binary
FROM golang:1.20 as builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY internal/controller/ internal/controller/

# Build
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o manager cmd/main.go

# Use Red Hat UBI base image
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest
ARG VERSION
LABEL name=pachyderm-operator \
      vendor='Pachyderm, Inc.' \
      version=$VERSION \
      release='beta' \
      description='Operator to manage Pachyderm instances' \
      summary='Pachyderm is the data foundation for Machine Learning. It combines Data Lineage with End-to-End Pipelines'

ENV USER_ID=1001
ADD LICENSE /licenses/apache2
ADD hack/charts /charts

WORKDIR /
COPY --from=builder /workspace/manager .
USER ${USER_ID}

ENTRYPOINT ["/manager"]
