SHELL := /bin/bash
.ONESHELL:

export CGO_ENABLED=1

export C_INCLUDE_PATH=$(DART_SDK_INCLUDE_DIR):$(abspath ./go/)

DART_SDK_INCLUDE_DIR = $(shell bash -c 'dirname `which flutter`')/cache/dart-sdk/include/

ALL_GOSRC = $(shell bash -O globstar -c 'echo go/**/*.go')
GOSRC = $(shell grep -Ril $(ALL_GOSRC) -e 'import "C"')

ALL_HSRC = $(shell bash -O globstar -c 'echo go/**/*.h')

BUILD_DIR = $(abspath ./build)
WORK_DIR = $(abspath .)

FAST = $(if $(fast),true,false)

build/include/go.h: $(GOSRC) $(ALL_HSRC)
	rm build -rf
	mkdir build/include/ -p
	entry_hfile=$(BUILD_DIR)/include/go.h
	cd go
	for file in $(GOSRC); do
		file=$${file#"go/"}
		hfile=$${file}.h
		hpath=$(BUILD_DIR)/include/$${hfile}
		mkdir -p $$(dirname $${hpath})
		go tool cgo -exportheader $${hpath} $${file}
		echo '#include "'$${hfile}'"' >> $${entry_hfile}
	done

	for file in $(ALL_HSRC); do
		file=$${file#"go/"}
		dest=$(BUILD_DIR)/include/$${file}
		mkdir -p $$(dirname $${dest})
		cp $${file} $${dest}
	done


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
	dart run test --reporter=expanded --debug --chain-stack-traces -j 1
endef

.PHONY: test_basic
test_basic:
	$(call integration_test,test_basic)

tidy_go = (cd $(WORK_DIR)/$1; go mod tidy; go fmt)
tidy_dart = (cd $(WORK_DIR)/$1; dart run import_sorter:main; dart fix --apply; dart format --fix .)

.PHONY: test_gen
test_gen:
	if ! $(FAST); then
		if ! go run github.com/hsfzxjy/dgo/dgo-gen tests/test_gen/go -i ../_build/ir.json; then
			exit 1
		fi
	fi
	$(call integration_test,test_gen)

.PHONY: run
run:
	cd $(WORK_DIR)/$(at)
	$(cmd)