.PHONY: depend clean test build tarball
.DEFAULT: build

GODEP=GOPATH="`godep path`:$(GOPATH)"

PLATFORMS := windows linux darwin
ARCH := amd64 386
PATH_SEP := /
BUILD_PAIRS := $(foreach p,$(PLATFORMS), \
	$(foreach a,$(ARCH),$(p)/$(a)) \
)
BUILD_DOCS := README.md LICENSE example_config.yaml

package: $(BUILD_PAIRS)

build: depend clean test
	@echo
	@echo "\033[32mBuilding ----> \033[m"
	$(GODEP) gox -os="$(PLATFORMS)" -arch="$(ARCH)" -output "build/{{.OS}}/{{.Arch}}/remote_syslog/remote_syslog"

clean:
	@echo
	@echo "\033[32mCleaning Build ----> \033[m"
	$(RM) -rf pkg/*
	$(RM) -rf build/*

test:
	@echo
	@echo "\033[32mTesting ----> \033[m"
	$(GODEP) go test ./...


depend:
	@echo
	@echo "\033[Checking Dependencies ----> \033[m"
	chmod +x ./build_deps.sh
	./build_deps.sh

$(BUILD_PAIRS): build
	@echo
	@echo "\033[32mPackaging ----> $@\033[m"
	$(eval PLATFORM := $(strip $(subst /, ,$(dir $@))))
	$(eval ARCH := $(notdir $@))
	mkdir pkg || echo
	cp $(BUILD_DOCS) build/$@/remote_syslog
	cd build/$@ && echo `pwd` && tar -cvzf ../../../pkg/remote_syslog_$(PLATFORM)_$(ARCH).tar.gz remote_syslog



