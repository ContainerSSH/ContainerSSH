export GO111MODULE=on
export GOOS?=$(shell uname | awk '{print tolower($$0)}')
GOVER=$(shell cat go.mod| grep -e "^go " | cut -d" " -f2)
GOFILES=$(shell find . -type f -name '*.go' -not -path "./.git/*")

fmt:
	go fmt ./...

fmtcheck:
	([ -z "$(shell gofmt -d $(GOFILES))" ]) || (echo "Source is unformatted, please execute make format"; exit 1)

vet:
	go vet ./...

test:
	go test -timeout 30s -covermode=atomic -coverprofile=cover.out ./...

build: fmtcheck vet test
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
