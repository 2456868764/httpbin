# Build the manager binary
FROM golang:1.20.1-alpine3.17 as builder

# Build Args
ARG TARGETOS
ARG TARGETARCH
ARG LDFLAGS
ARG PKGNAME
ARG BUILD
WORKDIR /app

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN if [[ "${BUILD}" != "CI" ]]; then go env -w GOPROXY=https://goproxy.io,direct; fi
RUN go env
RUN go mod download

# Copy the go source
COPY api api/
COPY pkg pkg/
COPY cmd cmd/

# Build
RUN env
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -ldflags="${LDFLAGS}" -a -o httpbin cmd/main.go

FROM alpine:3.17
WORKDIR /app
ARG PKGNAME
COPY --from=builder /app/httpbin .
CMD ["/app/httpbin"]
