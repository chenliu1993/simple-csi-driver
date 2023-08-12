.DEFAULT_GOAL:=help
SHELL := /usr/bin/env bash

# Repo
REPO_NAME := "github.com/chenliu1993/simple-csi-driver"
ROOT_DIR := $(shell git rev-parse --show-toplevel)

# Git
GITCOMMIT ?= `git rev-parse HEAD`
BUILD_DATE = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Tools
TOOLS_DIR := ${ROOT_DIR}/hack

# GO
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
GOPATH ?= $(shell go env GOPATH)
GOBIN ?= $(GOPATH)/bin
GO111MODOLE := on
export GO111MODULE GOPATH GOBIN

## GOLANGCI_LINT
GOLANGCI_LINT_VERSION := v1.54.0

.PHONY: help
help: #### display help
	@awk 'BEGIN {FS = ":.*## "; printf "\nTargets:\n"} /^[a-zA-Z_-]+:.*?#### / { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@awk 'BEGIN {FS = ":.* ## "; printf "\n  \033[1;32mBuild targets\033[36m\033[0m\n  \033[0;37mTargets for building and/or installing CLI plugins on the system.\n  Append \"ENVS=<os-arch>\" to the end of these targets to limit the binaries built.\n  e.g.: make build-all-tanzu-cli-plugins ENVS=linux-amd64  \n  List available at https://github.com/golang/go/blob/master/src/go/build/syslist.go\033[36m\033[0m\n\n"} /^[a-zA-Z_-]+:.*? ## / { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
##### GLOBAL

.PHONY: tidy
tidy: #### do a tidy
	GOPRIVATE=${REPO_NAME} go mod tidy

.PHONY: lint
lint:  #### check go code style
	${TOOLS_DIR}/bin/golangci-lint run ./...

.PHONY: fmt
fmt: #### format go code
	go fmt ./...

.PHONY: vet
vet: #### vet go code
	go vet ./...

.PHONY: vendor
vendor: #### vendor go code
	go mod vendor

.PHONY: update-modules
update-modules: tidy vendor#### update go modules

.PHONY: tools-dir
tools-dir: #### create tools dir
	rm -rf ${TOOLS_DIR}/bin && mkdir -p ${TOOLS_DIR}/bin

.PHONY: golangci-lint
golangci-lint: #### install golangci-lint
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${TOOLS_DIR}/bin ${GOLANGCI_LINT_VERSION}

.PHONY: unittest
unittest: #### run unittest
	go test -v $(shell go list ./... | grep -v /test/) -coverprofile=coverage.txt -covermode=count

# .PHONY: prepare-sanity
# prepare-sanity: #### prepare sanitytest
# 	docker run -d --name nfs --privileged -p 2049:2049 -v "$(pwd)"/nfsshare:/nfsshare -e SHARED_DIRECTORY=/nfsshare itsthenetwork/nfs-server-alpine:latest
.PHONY: sanitytest
sanitytest: #### run sanitytest
	go test -v ${ROOT_DIR}/test/sanity

.PHONY: build
build: #### build binary
	GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 \
		go build -mod=vendor -a -ldflags "-X github.com/chenliu1993/simple-csi-driver/internal/nfs:nfsDriverVersion=${GITCOMMIT} -s -w -extldflags "-static"" \
		-o simple-csi-driver ${ROOT_DIR}

.PHONY: build-image
build-image: #### build a iamge
	docker build . -t cl20192019/simple-csi-driver:${GITCOMMIT}

.PHONY: push-image
push-image: #### push a image
	docker push cl20192019/simple-csi-driver:${GITCOMMIT}