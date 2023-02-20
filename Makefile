.SILENT:

CURR_DIR=$(dir $(abspath $(lastword $(MAKEFILE_LIST))))

PROTO_FILES=$(wildcard $(CURR_DIR)*/*/*.proto)
PROTO_FILES_BASE=$(basename $(PROTO_FILES))
PROTO_GO_FILES=$(addsuffix .pb.go, $(PROTO_FILES_BASE))

PROTOC=protoc
PROTOC_OPTS=--proto_path=$(CURR_DIR) --go_out=$(CURR_DIR) --go_opt=paths=source_relative

vmdiff: proto fmt
	go build

.PHONY: proto
proto: $(PROTO_GO_FILES)

$(PROTO_GO_FILES): $(PROTO_FILES)
	$(PROTOC) $(PROTOC_OPTS) $(@:.pb.go=.proto)

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test: proto
	go test `go list ./... | grep -v proto | grep -v testhelper` $(testargs)

.PHONY: install
install: fmt proto
	go install

.PHONY: uninstall
uninstall:
	go clean -i

.PHONY: coverage
coverage:
	$(MAKE) test testargs="-coverprofile=coverage.out" && \
		go tool cover -html=coverage.out

.PHONY: all
all: vmdiff test

.PHONY: clean
clean:
	-rm -f $(PROTO_GO_FILES) coverage.out
	go clean
	go clean -testcache
