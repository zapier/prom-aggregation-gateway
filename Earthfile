VERSION 0.6

ARG image_name="prom-aggregation-gateway"
ARG token=""

ARG commitSHA=""
ARG version="dev"
ARG PKG_PATH="github.com/zapier/prom-aggregation-gateway"

ARG CHART_RELEASER_VERSION="1.4.1"
ARG CHART_TESTING_VERSION="3.7.1"
ARG GITHUB_CLI_VERSION="2.20.2"
ARG GOLANG_VERSION="1.19.3"
ARG HELM_UNITTEST_VERSION="0.2.8"
ARG KUBECONFORM_VERSION="0.5.0"
ARG STATICCHECK_VERSION="0.3.3"

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

release:
    BUILD +release-binaries
    BUILD +build-image

go-deps:
    FROM "golang:${GOLANG_VERSION}-bullseye"

    WORKDIR /src
    COPY go.mod go.sum /src
    RUN go mod download

build-binary:
    FROM +go-deps

    WORKDIR /src
    COPY . /src
    RUN go build -ldflags "-X ${PKG_PATH}/config.Version=${version} -X ${PKG_PATH}/config.CommitSHA=${commitSHA}" -o prom-aggregation-gateway .

    SAVE ARTIFACT ./prom-aggregation-gateway

build-image:
    FROM bullseye
    COPY +build-binary/prom-aggregation-gateway .
    ENV GIN_MODE=release
    USER 65534
    ENTRYPOINT ["/prom-aggregation-gateway"]
    SAVE IMAGE --push ${image_name}:${version}

build-rpm:
    FROM centos:centos7

    RUN yum install -y wget

    ARG NFPM_VERSION="2.32.0"
    RUN \
        wget \
            https://github.com/goreleaser/nfpm/releases/download/v${NFPM_VERSION}/nfpm_${NFPM_VERSION}_Linux_x86_64.tar.gz \
            --output-document nfpm.tgz \
        && tar xvf nfpm.tgz \
        && mv ./nfpm /bin \
        && nfpm --version

    WORKDIR /usr/src/pag

    ENV VERSION=$version
    ENV RELEASE=0
    COPY spec/pag.yml .
    COPY +build-binary/prom-aggregation-gateway .
    RUN \
        mkdir ./dist/ \
        && nfpm pkg \
            --packager rpm \
            --config ./pag.yml \
            --target ./dist/prom-aggregation-gateway-${VERSION}.rpm
    SAVE ARTIFACT ./dist/ AS LOCAL ./dist/

test-rpm:
    FROM centos:centos7

    ENV VERSION=$version
    ENV RELEASE=0

    COPY +build-rpm/dist/prom-aggregation-gateway-${VERSION}.rpm pag.rpm
    RUN rpm -i pag.rpm && prom-aggregation-gateway version

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
    FROM bullseye

    COPY . /src
    WORKDIR /src
    COPY +build-binaries/_dist dist
    COPY +build-rpm/dist dist

    # install github cli
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
    RUN FILE=staticcheck.tgz \
        && URL=https://github.com/dominikh/go-tools/releases/download/v${STATICCHECK_VERSION}/staticcheck_linux_amd64.tar.gz \
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
    FROM quay.io/helmpack/chart-testing:v${CHART_TESTING_VERSION}

    # install kubeconform
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

    RUN apk add --no-cache bash git \
        && helm plugin install --version "${HELM_UNITTEST_VERSION}" https://github.com/quintush/helm-unittest \
        && helm unittest --help

    # actually lint the chart
    WORKDIR /src
    COPY . /src
    RUN git fetch --prune --unshallow | true
    RUN ct --config ./.github/ct.yaml lint ./charts

build-helm:
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
