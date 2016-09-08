#!/bin/sh

BUILDPATH="build/remote_syslog2"

set -e

mkdir -p $BUILDPATH

go build -o $BUILDPATH/remote_syslog .
cp README.md LICENSE example_config.yml $BUILDPATH

cd $BUILDPATH/..
rm -f remote_syslog2.tar.gz
tar -czf remote_syslog2.tar.gz `basename $BUILDPATH`
rm -r remote_syslog2
