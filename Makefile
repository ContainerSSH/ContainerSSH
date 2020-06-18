export GO111MODULE=on
export GOOS?=$(shell uname | awk '{print tolower($$0)}')
export GOLICENSE_VERSION=0.2.0
export CGO_ENABLED=0
GOFMT_LINES := $(shell gofmt -l . | wc -l)
GOVER=$(shell cat go.mod| grep -e "^go " | cut -d" " -f2)

golicense:
	cd build && curl https://github.com/mitchellh/golicense/releases/download/v${GOLICENSE_VERSION}/golicense_${GOLICENSE_VERSION}_linux_x86_64.tar.gz | tar -xvzf && chmod +x golicense

fmt:
	go fmt ./...

fmtcheck:
	ifneq "$(GOFMT_LINES)" "0"
		$(error $(GOFMT_LINES) files are not properly formatted. Please run gofmt)
	fi

vet:
	go vet ./...

test:
	go test -timeout 30s -covermode=atomic -coverprofile=cover.out ./...

licensecheck: golicense
	# Dummy build for license check
	GOOS=linux GOARCH=x86_64 go build -o build/containerssh cmd/containerssh/main.go
	build/golicense golicense.json build/containerssh >license-check-results.txt
	rm build/containerssh

build: fmtcheck vet test.go licensecheck
	go build -o build/containerssh cmd/containerssh/main.go
	go build -o build/testAuthConfigServer cmd/testAuthConfigServer/main.go

build-docker:
	@#USER_NS='-u $(shell id -u $(whoami)):$(shell id -g $(whoami))'
	docker run \
		--rm \
		${USER_NS} \
		-v "${PWD}":/go/src/github.com/janoszen/containerssh \
		-w /go/src/github.com/janoszen/containerssh \
		-e GO111MODULE=on \
		-e GOOS=${GOOS} \
		golang:${GOVER} \
		make build
