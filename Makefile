.SILENT:

CURR_DIR=$(dir $(abspath $(lastword $(MAKEFILE_LIST))))

PROTO_FILES=$(wildcard $(CURR_DIR)*/*.proto)
PROTO_FILES_BASE=$(basename $(PROTO_FILES))
GO_FILES=$(addsuffix .pb.go, $(PROTO_FILES_BASE))

PROTOC=protoc
PROTOC_OPTS=--proto_path=$(CURR_DIR) --go_out=$(CURR_DIR) --go_opt=paths=source_relative

vmdiff: fmt proto
	go build

proto: $(GO_FILES)

$(GO_FILES): $(PROTO_FILES)
	$(PROTOC) $(PROTOC_OPTS) $(@:.pb.go=.proto)

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: test
test:
	go test ./...

.PHONY: all
all: vmdiff test

.PHONY: clean
clean:
	-rm -f vmdiff $(GO_FILES)
