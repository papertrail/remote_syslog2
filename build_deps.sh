#!/usr/bin/env sh

type go >/dev/null 2>&1 || {
  yellow "Go is required to build this application";
  yellow "If you are using homebrew on OSX, run"
  code "brew install go --cross-compile-all"
  exit 1;
}

if [[ -z $GOPATH ]]; then
  yellow "GOPATH is not set. This means that you do not have go setup properly on this machine"
  code "mkdir ~/gocode"
  code "echo 'export GOPATH=~/gocode' >> ~/.bash_profile"
  code "echo 'export PATH=\"$GOPATH/bin:$PATH\"' >> ~/.bash_profile"
  code "source ~/.bash_profile"
  exit 1
fi

type godep >/dev/null 2>&1 || {
  yellow "Godep is not installed. See https://github.com/kr/godep";
  code" go get github.com/kr/godep"
  exit 1;
}

type gox >/dev/null 2>&1 || {
  yellow "Gox is not installed. See https://github.com/mitchellh/gox";
  code "go get github.com/mitchellh/gox"
  exit 1;
}
