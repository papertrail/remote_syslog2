#!/bin/sh

set -e

mkdir -p build/remote_syslog2

godep go build -o build/remote_syslog2/remote_syslog2 .
cp README.md LICENSE example_config.json build/remote_syslog2

cd pkg
rm -f remote_syslog2.tar.gz
tar -czf remote_syslog2.tar.gz remote_syslog2
rm -r remote_syslog2
