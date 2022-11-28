SHELL := /bin/bash
.ONESHELL:

export CGO_ENABLED=1

export C_INCLUDE_PATH=$(DART_SDK_INCLUDE_DIR):$(abspath ./go/)

DART_SDK_INCLUDE_DIR = $(shell bash -c 'dirname `which flutter`')/cache/dart-sdk/include/

GOSRC = $(abspath $(shell grep -Ril go/*.go -e 'import "C"'))

BUILD_DIR = $(abspath ./build)
WORK_DIR = $(abspath .)

build/include/go.h: $(GOSRC)
	rm build -rf
	mkdir build/include/ -p
	cd go
	go tool cgo -exportheader $(BUILD_DIR)/include/go.h $(GOSRC)

.PHONY: go-headers
go-headers: build/include/go.h

dart/lib/dgo_binding.dart: build/include/go.h $(wildcard go/*.h) dart/pubspec.yaml
	cd dart
	dart run ffigen

.PHONY: ffigen
ffigen: dart/lib/dgo_binding.dart

.PHONY: test
test:
	cd tests/test_basic/go
	make
	cd ../dart
	dart run ffigen
	dart run test --reporter=expanded --debug

tidy_go = (cd $(WORK_DIR)/$1; go mod tidy; go fmt)
tidy_dart = (cd $(WORK_DIR)/$1; dart run import_sorter:main; dart fix --apply)

.PHONY: tidy
tidy:
	$(call tidy_go,go)
	$(call tidy_go,dgo-gen)

	$(call tidy_dart,dart)
	$(call tidy_dart,tests/test_basic/dart)
	$(call tidy_dart,dgo-gen-dart)

.PHONY: test_gen
test_gen:
	go run github.com/hsfzxjy/dgo/dgo-gen tests/test_gen/go
	cd tests/test_gen/go
	make
	cd ../dart
	dart run test --reporter=expanded --debug --chain-stack-traces

.PHONY: run
run:
	cd $(WORK_DIR)/$(at)
	$(cmd)