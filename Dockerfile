FROM golang:1.19 AS builder

# go install github.com/grpc-ecosystem/grpc-health-probe@latest
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.11 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

RUN apt-get update
RUN apt-get install ca-certificates -y
RUN update-ca-certificates

ENV GOPATH=/go
ENV PATH=${GOPATH}/bin:$PATH

WORKDIR "/build/src"

COPY go.mod .
COPY go.sum .
RUN --mount=type=cache,target=/go/pkg/mod go mod download
RUN --mount=type=cache,target=/go/pkg/mod go mod verify

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod CGO_ENABLED=0 GOOS=linux go build -tags release -ldflags '-s -d -w' -o /build/dist/grpc_reflection /build/src/cmd/reflection

#FROM scratch
#COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
#COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
FROM alpine:latest

COPY --from=builder /bin/grpc_health_probe /bin/grpc_health_probe
COPY --from=builder "/build/dist/" /

ENTRYPOINT ["/grpc_reflection"]
