COREDNS_VERSION:=1.12.4

build/coredns.tar.gz:
	mkdir build
	wget -O $@ https://github.com/coredns/coredns/archive/refs/tags/v$(COREDNS_VERSION).tar.gz

build/coredns: build/coredns.tar.gz
	tar -xzf $^ -o -C build
	mv build/coredns-* build/coredns

build/coredns/plugin/%: plugin/% build/coredns
	cp -r $< build/coredns/plugin
	sed s:github.com/sneakybugs/corewarden:github.com/coredns: -i $@$/**/*.go

build/coredns/plugin.cfg: plugin.cfg
	cp $^ build/coredns

srcplugins := $(shell echo plugin/*)
plugins := $(patsubst %,build/coredns/%,$(srcplugins))

coredns: $(plugins) build/coredns/plugin.cfg
	go mod tidy -C build/coredns
	go generate -C build/coredns
	go build -C build/coredns -o ../../$@

.PHONY: clean
clean:
	rm -rf build
	rm coredns
