SHELL := /bin/bash
.ONESHELL:

export CGO_ENABLED=1

export C_INCLUDE_PATH=$(DART_SDK_INCLUDE_DIR):$(abspath ./go/)

DART_SDK_INCLUDE_DIR = $(shell bash -c 'dirname `which flutter`')/cache/dart-sdk/include/

GOSRC = $(abspath $(shell grep -Ril go/*.go -e 'import "C"'))

BUILD_DIR = $(abspath ./build)

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
	cd tests/dummyso/
	make
	cd ../../dart
	export LD_LIBRARY_PATH=$(BUILD_DIR)
	dart run test --reporter=expanded --debug

.PHONY: tidy
tidy:
	cd go
	go mod tidy

	cd ../dgo-gen/
	go mod tidy

	cd ../dart/
	dart pub run import_sorter:main
	dart fix --apply

	cd ../dgo-gen-dart/
	dart pub run import_sorter:main
	dart fix --apply