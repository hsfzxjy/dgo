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

.PHONY: tidy
tidy:
	$(call tidy_go,go)
	$(call tidy_go,tests/test_basic/go)
	$(call tidy_go,dgo-gen)
	$(call tidy_go,tests/test_gen/go)

	$(call tidy_dart,dart)
	$(call tidy_dart,tests/test_basic/dart)
	$(call tidy_dart,dgo-gen-dart)
	$(call tidy_dart,tests/test_gen/dart)

define integration_test
	cd $(WORK_DIR)/tests/$1/go; \
	make; \
	cd ../dart; \
	dart run test --reporter=expanded --debug --chain-stack-traces
endef

.PHONY: test_basic
test_basic:
	$(call integration_test,test_basic)

tidy_go = (cd $(WORK_DIR)/$1; go mod tidy; go fmt)
tidy_dart = (cd $(WORK_DIR)/$1; dart run import_sorter:main; dart fix --apply)

.PHONY: test_gen
test_gen:
	if ! go run github.com/hsfzxjy/dgo/dgo-gen tests/test_gen/go -i ../_build/ir.json; then
		exit 1
	fi
	$(call integration_test,test_gen)

.PHONY: run
run:
	cd $(WORK_DIR)/$(at)
	$(cmd)