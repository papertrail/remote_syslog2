.PHONY: depend clean test build tarball
.DEFAULT: build

GODEP=GOPATH="`godep path`:$(GOPATH)"

X86_PLATFORMS := windows linux
X64_PLATFORMS := windows linux darwin

BUILD_PAIRS := $(foreach p,$(X86_PLATFORMS), $(p)/386 )
BUILD_PAIRS += $(foreach p,$(X64_PLATFORMS), $(p)/amd64 )

BUILD_DOCS := README.md LICENSE example_config.yml

package: $(BUILD_PAIRS)

build: depend clean test
	@echo
	@echo "\033[32mBuilding ----> \033[m"
	$(GODEP) gox -os="$(X64_PLATFORMS)" -arch="amd64" -output "build/{{.OS}}/{{.Arch}}/remote_syslog/remote_syslog"
	$(GODEP) gox -os="$(X86_PLATFORMS)" -arch="386" -output "build/{{.OS}}/{{.Arch}}/remote_syslog/remote_syslog"


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
	@echo "\033[32mChecking Dependencies ----> \033[m"

ifndef GOPATH
	@echo "\033[1;33mGOPATH is not set. This means that you do not have go setup properly on this machine\033[m"
	@echo "$$ mkdir ~/gocode";
	@echo "$$ echo 'export GOPATH=~/gocode' >> ~/.bash_profile";
	@echo "$$ echo 'export PATH=\"\$$GOPATH/bin:\$$PATH\"' >> ~/.bash_profile";
	@echo "$$ source ~/.bash_profile";
	exit 1;
endif

	type go >/dev/null 2>&1|| { \
	  echo "\033[1;33mGo is required to build this application\033[m"; \
	  echo "\033[1;33mIf you are using homebrew on OSX, run\033[m"; \
	  echo "$$ brew install go --cross-compile-all"; \
	  exit 1; \
	}

	type godep >/dev/null 2>&1|| { \
	  echo "\033[1;33mGodep is not installed. See https://github.com/tools/godep\033[m"; \
	  echo "$$ go get github.com/tools/godep"; \
	  exit 1; \
	}

	type gox >/dev/null 2>&1 || { \
	  echo "\033[1;33mGox is not installed. See https://github.com/mitchellh/gox\033[m"; \
	  echo "$$ go get github.com/mitchellh/gox"; \
	  exit 1; \
	}



$(BUILD_PAIRS): build
	@echo
	@echo "\033[32mPackaging ----> $@\033[m"
	$(eval PLATFORM := $(strip $(subst /, ,$(dir $@))))
	$(eval ARCH := $(notdir $@))
	mkdir pkg || echo
	cp $(BUILD_DOCS) build/$@/remote_syslog
	cd build/$@ && echo `pwd` && tar -cvzf ../../../pkg/remote_syslog_$(PLATFORM)_$(ARCH).tar.gz remote_syslog



