export CC=$(shell which gcc)
export CXX=$(shell which g++)

.PHONY: all
all:
	mkdir -p ../_build/
	go build -buildmode=c-shared -o ../_build/libtest_gen.so
