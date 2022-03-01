SHELL = /bin/bash
NAME=$(shell basename $(CURDIR))
VERSION=$(shell git describe --tags `git rev-list --tags --max-count=1`|| echo "unkown version")
RELEASE=$(shell git describe --abbrev=1 --tags  | awk -F- '{if (NF > 1) print $$2; else print 0}')
BUILDTIME=$(shell date '+%Y-%m-%d %H:%M:%S %z')
Commit=$(shell git rev-parse --short HEAD)
TARGET=${NAME}-${SHORT_VERSION}
ifeq (${DF},)
        DF := build/Dockerfile
endif

GOBUILD=CGO_CFLAGS=-Wno-undef-prefix \
        go build -ldflags '-w -s\
        -X "github.com/NetEase-Media/ngo/g.Version=$(VERSION)" \
        -X "github.com/NetEase-Media/ngo/g.BuildTime=$(BUILDTIME)" \
        -X "github.com/NetEase-Media/ngo/g.Commit=$(Commit)" \
        -X "github.com/NetEase-Media/ngo/g.ProgName=$(NAME)"'
SOURCES="./cmd/"

SCRIPT_DIR = $(shell pwd)/etc/script
PKG_LIST   = $(shell go list ./... | grep -v /vendor/ | grep -v /examples)

export GO111MODULE=on
export GOPROXY=https://goproxy.io
# exported to submakes
export

all: gobuild

gobuild:
	$(GOBUILD) -o $(NAME)  $(SOURCES)

test:
	go test ./...

bench:
	go test ./... -bench=.

clean:
	git clean -xdf

lint:
	golangci-lint run

run:
	/$(NAME) -c ./app.yaml

docker:
	docker build -f ${DF} -t $(NAME):$(VERSION) ./

mod:
	go mod tidy

code_coverage: dep ## Generate global code coverage report
	sh ${SCRIPT_DIR}/coverage.sh

code_coverage_html: dep
	sh ${SCRIPT_DIR}/coverage.sh html;

race_detector: dep ## Run data race detector
	go test -gcflags=-l -race -short ${PKG_LIST}

unit_tests: dep ## Run unittests
	go test -gcflags=-l -v ${PKG_LIST}

.PHONY: all clean