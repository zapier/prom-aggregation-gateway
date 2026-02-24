FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.25.7-alpine AS builder

ARG VERSION=dev
ARG COMMIT_SHA=""
ARG PKG_PATH=github.com/zapier/prom-aggregation-gateway
ARG TARGETOS
ARG TARGETARCH

WORKDIR /src
COPY go.mod go.sum /src/
RUN go mod download
COPY . /src
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
  -ldflags "-X ${PKG_PATH}/config.Version=${VERSION} -X ${PKG_PATH}/config.CommitSHA=${COMMIT_SHA}" \
  -o prom-aggregation-gateway \
  .

FROM --platform=${BUILDPLATFORM:-linux/amd64} gcr.io/distroless/static:nonroot

COPY --from=builder /src/prom-aggregation-gateway /prom-aggregation-gateway
ENV GIN_MODE=release
USER 65532:65532
ENTRYPOINT ["/prom-aggregation-gateway"]
