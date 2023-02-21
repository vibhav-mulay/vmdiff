.SILENT:

CURR_DIR=$(dir $(abspath $(lastword $(MAKEFILE_LIST))))

PROTO_FILES=$(wildcard $(CURR_DIR)*/*.proto)
PROTO_FILES_BASE=$(basename $(PROTO_FILES))
PROTO_GO_FILES=$(addsuffix .pb.go, $(PROTO_FILES_BASE))

PROTOC=protoc
PROTOC_OPTS=--proto_path=$(CURR_DIR) --go_out=$(CURR_DIR) --go_opt=paths=source_relative

vmdiff-cli: proto fmt
	@$(MAKE) -C vmdiff-cli

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
	@$(MAKE) -C vmdiff-cli install
	go install

.PHONY: uninstall
uninstall:
	@$(MAKE) -C vmdiff-cli uninstall
	go clean -i

.PHONY: coverage
coverage:
	$(MAKE) test testargs="-coverprofile=coverage.out" && \
		go tool cover -html=coverage.out

.PHONY: all
all: vmdiff-cli install test

.PHONY: clean
clean:
	@$(MAKE) -C vmdiff-cli clean
	-rm -f $(PROTO_GO_FILES) coverage.out
	go clean -testcache
