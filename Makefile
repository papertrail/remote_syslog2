include packaging/Makefile.packaging

.PHONY: depend clean test build tarball

GOFLAGS="-trimpath"
GOLDFLAGS="-X main.Version=$(PACKAGE_VERSION)"

X86_PLATFORMS := windows linux freebsd
X64_PLATFORMS := windows linux freebsd darwin
ARM_PLATFORMS := linux

BUILD_PAIRS := $(foreach p,$(X86_PLATFORMS), $(p)/i386 )
BUILD_PAIRS += $(foreach p,$(X64_PLATFORMS), $(p)/amd64 )
BUILD_PAIRS += $(foreach p,$(ARM_PLATFORMS), $(p)/armhf )
BUILD_PAIRS += $(foreach p,$(ARM_PLATFORMS), $(p)/arm64 )

BUILD_DOCS := README.md LICENSE example_config.yml

package: $(BUILD_PAIRS)


build: depend clean test
	@echo
	@echo "\033[32mBuilding ----> \033[m"
	GOFLAGS=$(GOFLAGS) gox -ldflags=$(GOLDFLAGS) -os="$(X64_PLATFORMS)" -arch="amd64" -output "build/{{.OS}}/amd64/remote_syslog/remote_syslog"
	GOFLAGS=$(GOFLAGS) gox -ldflags=$(GOLDFLAGS) -os="$(X86_PLATFORMS)" -arch="386" -output "build/{{.OS}}/i386/remote_syslog/remote_syslog"
	GOFLAGS=$(GOFLAGS) gox -ldflags=$(GOLDFLAGS) -os="$(ARM_PLATFORMS)" -arch="arm" -output "build/{{.OS}}/armhf/remote_syslog/remote_syslog"
	GOFLAGS=$(GOFLAGS) gox -ldflags=$(GOLDFLAGS) -os="$(ARM_PLATFORMS)" -arch="arm64" -output "build/{{.OS}}/arm64/remote_syslog/remote_syslog"


clean:
	@echo
	@echo "\033[32mCleaning Build ----> \033[m"
	$(RM) -rf pkg/*
	$(RM) -rf build/*
	$(RM) -rf tmp/*


test:
	@echo
	@echo "\033[32mTesting ----> \033[m"
	go test -v -race ./...


depend:
	@echo
	@echo "\033[32mChecking Dependencies ----> \033[m"

ifndef PACKAGE_VERSION
	@echo "\033[1;33mPACKAGE_VERSION is not set. In order to build a package I need PACKAGE_VERSION=n\033[m"
	exit 1;
endif

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

	type gox >/dev/null 2>&1 || { \
	  echo "\033[1;33mGox is not installed. See https://github.com/mitchellh/gox\033[m"; \
	  echo "$$ go install github.com/mitchellh/gox"; \
	  exit 1; \
	}

	gem list | grep fpm >/dev/null 2>&1 || { \
	  echo "\033[1;33mfpm is not installed. See https://github.com/jordansissel/fpm\033[m"; \
	  echo "$$ gem install fpm"; \
	  exit 1; \
	}

	type rpmbuild >/dev/null 2>&1 || { \
	  echo "\033[1;33mrpmbuild is not installed. See the package for your distribution\033[m"; \
	  exit 1; \
	}


$(BUILD_PAIRS): build
	@echo
	@echo "\033[32mPackaging ----> $@\033[m"
	$(eval PLATFORM := $(strip $(subst /, ,$(dir $@))))
	$(eval ARCH := $(notdir $@))
	mkdir pkg || echo
	cp $(BUILD_DOCS) build/$@/remote_syslog

	if [ "$(PLATFORM)" = "linux" ]; then\
		mkdir -p pkg/tmp/etc/init.d;\
		mkdir -p pkg/tmp/usr/local/bin;\
		cp -f example_config.yml pkg/tmp/etc/log_files.yml;\
		cp -f packaging/linux/remote_syslog.initd pkg/tmp/etc/init.d/remote_syslog;\
		cp -f build/$@/remote_syslog/remote_syslog pkg/tmp/usr/local/bin;\
		(cd pkg && \
		fpm \
		  -s dir \
		  -C tmp \
		  -t deb \
		  -n $(PACKAGE_NAME) \
		  -v $(PACKAGE_VERSION) \
		  --vendor $(PACKAGE_VENDOR) \
		  --license $(PACKAGE_LICENSE) \
		  -a $(ARCH) \
		  -m $(PACKAGE_CONTACT) \
		  --description $(PACKAGE_DESCRIPTION) \
		  --url $(PACKAGE_URL) \
		  --before-remove ../packaging/linux/deb/prerm \
		  --after-install ../packaging/linux/deb/postinst \
		  --config-files etc/log_files.yml \
		  --config-files etc/init.d/remote_syslog usr/local/bin/remote_syslog etc/log_files.yml etc/init.d/remote_syslog && \
		fpm \
		  -s dir \
		  -C tmp \
		  -t rpm \
		  -n $(PACKAGE_NAME) \
		  -v $(PACKAGE_VERSION) \
		  --vendor $(PACKAGE_VENDOR) \
		  --license $(PACKAGE_LICENSE) \
		  -a $(ARCH) \
		  -m $(PACKAGE_CONTACT) \
		  --description $(PACKAGE_DESCRIPTION) \
		  --url $(PACKAGE_URL) \
		  --before-remove ../packaging/linux/rpm/preun \
		  --after-install ../packaging/linux/rpm/post \
		  --config-files etc/log_files.yml \
		  --config-files etc/init.d/remote_syslog \
		  --rpm-os linux usr/local/bin/remote_syslog etc/log_files.yml etc/init.d/remote_syslog );\
		rm -R -f pkg/tmp;\
	fi

	cd build/$@ && echo `pwd` && tar -cvzf ../../../pkg/remote_syslog_$(PLATFORM)_$(ARCH).tar.gz remote_syslog
