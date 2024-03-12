VERSION 0.8

ARG --global token=""

ARG --global commitSHA=""
ARG --global version="dev"
ARG --global PKG_PATH="github.com/zapier/prom-aggregation-gateway"

ARG --global ALPINE_VERSION="3.18"
ARG --global --required GOLANG_VERSION

test:
    BUILD +ci-golang

ci-golang:
    BUILD +lint-golang
    BUILD +test-golang

ci-helm:
    BUILD +test-helm

build:
    BUILD +build-image
    BUILD +build-helm

release-multiarch:
    BUILD +release-binaries
    BUILD +build-image-multiarch

release:
    BUILD +release-binaries
    BUILD +build-image

go-deps:
    FROM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION}

    WORKDIR /src
    COPY go.mod go.sum /src
    RUN go mod download

build-binary:
    FROM +go-deps

    WORKDIR /src
    COPY . /src
    RUN go build -ldflags "-X ${PKG_PATH}/config.Version=${version} -X ${PKG_PATH}/config.CommitSHA=${commitSHA}" -o prom-aggregation-gateway .

    SAVE ARTIFACT ./prom-aggregation-gateway

build-image-multiarch:
    BUILD --platform=linux/arm64 --platform=linux/amd64 +build-image

build-image:
    FROM alpine:${ALPINE_VERSION}
    COPY +build-binary/prom-aggregation-gateway .
    ENV GIN_MODE=release
    USER 65534
    ENTRYPOINT ["/prom-aggregation-gateway"]

    ARG image_name="prom-aggregation-gateway"
    SAVE IMAGE --push ${image_name}:${version}
    SAVE IMAGE --push ${image_name}:latest

continuous-deploy:
    BUILD +build-helm

build-binaries:
    FROM golang:${GOLANG_VERSION}

    WORKDIR /src

    RUN go install github.com/mitchellh/gox@latest

    COPY go.mod go.sum /src
    RUN go mod download

    COPY . /src

    RUN \
        GOFLAGS="-trimpath" \
        GO111MODULE=on \
        CGO_ENABLED=0 \
        gox \
            -parallel=3 \
            -ldflags "-X ${PKG_PATH}/config.Version=${version}" \
            -ldflags "-X ${PKG_PATH}/config.CommitSHA=${commitSHA}" \
            -output="_dist/prom-aggregation-gateway-${version}-{{.OS}}-{{.Arch}}" \
            -osarch='darwin/amd64 darwin/arm64 linux/amd64 linux/386 linux/arm linux/arm64 linux/ppc64le linux/s390x windows/amd64' \
            .

    SAVE ARTIFACT _dist AS LOCAL ./dist

release-binaries:
    FROM alpine:${ALPINE_VERSION}

    COPY . /src
    WORKDIR /src
    COPY +build-binaries/_dist dist

    # install github cli
    ARG --required GITHUB_CLI_VERSION
    RUN FILE=ghcli.tgz \
        && URL=https://github.com/cli/cli/releases/download/v${GITHUB_CLI_VERSION}/gh_${GITHUB_CLI_VERSION}_linux_amd64.tar.gz \
        && wget ${URL} \
            --output-document ${FILE} \
        && tar \
            --extract \
            --verbose \
            --directory /usr \
            --strip-components=1 \
            --file ${FILE} \
        && gh version

    RUN apk add --no-cache git

    ENV GH_TOKEN $token
    RUN --push gh release upload ${version} ./dist/*

lint-golang:
    FROM +go-deps

    # install staticcheck
    ARG --required STATICCHECK_VERSION
    RUN FILE=staticcheck.tgz \
        && URL=https://github.com/dominikh/go-tools/releases/download/${STATICCHECK_VERSION}/staticcheck_linux_amd64.tar.gz \
        && wget ${URL} \
            --output-document ${FILE} \
        && tar \
            --extract \
            --verbose \
            --directory /bin \
            --strip-components=1 \
            --file ${FILE} \
        && staticcheck -version

    ENV CGO_ENABLED=0
    COPY . /src
    RUN staticcheck ./...

test-golang:
    FROM +go-deps

    COPY . /src

    ENV CGO_ENABLED=0
    RUN go test .

test-helm:
    ARG --required CHART_TESTING_VERSION

    FROM quay.io/helmpack/chart-testing:v${CHART_TESTING_VERSION}

    # install kubeconform
    ARG --required KUBECONFORM_VERSION
    RUN FILE=kubeconform.tgz \
        && URL=https://github.com/yannh/kubeconform/releases/download/v${KUBECONFORM_VERSION}/kubeconform-linux-amd64.tar.gz \
        && wget ${URL} \
            --output-document ${FILE} \
        && tar \
            --extract \
            --verbose \
            --directory /bin \
            --file ${FILE} \
        && kubeconform -v

    ARG HELM_UNITTEST_VERSION="0.2.8"
    RUN apk add --no-cache bash git \
        && helm plugin install --version "${HELM_UNITTEST_VERSION}" https://github.com/quintush/helm-unittest \
        && helm unittest --help

    # actually lint the chart
    WORKDIR /src
    COPY . /src
    RUN git fetch --prune --unshallow | true
    RUN ct --config ./.github/ct.yaml lint ./charts

build-helm:
    ARG --required CHART_RELEASER_VERSION

    FROM quay.io/helmpack/chart-releaser:v${CHART_RELEASER_VERSION}

    ARG token

    WORKDIR /src
    COPY . /src

    RUN cr --config .github/cr.yaml package charts/*
    SAVE ARTIFACT .cr-release-packages/ AS LOCAL ./dist

    RUN mkdir -p .cr-index
    RUN git config --global user.email "opensource@zapier.com"
    RUN git config --global user.name "Open Source at Zapier"
    RUN git fetch --prune --unshallow | true

    RUN --push cr --config .github/cr.yaml upload --token $token --skip-existing
    RUN --push cr --config .github/cr.yaml index --token $token --push
