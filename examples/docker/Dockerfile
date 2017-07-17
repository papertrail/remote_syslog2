FROM alpine:3.6
#MAINTAINER

# every system is different - every package unique
# remote_syslog2 works with docker - remote_syslog2 breaks with docker

# tested with v0.19 release - hardcoded -pull precompiled binares OR compile in dockerfile (>image)
ARG VERSION

RUN apk update \
  && apk add ca-certificates wget \
  && update-ca-certificates 

#build arg used to pull desired versions - docker cli 

RUN wget https://github.com/papertrail/remote_syslog2/releases/download/$VERSION/remote_syslog_linux_amd64.tar.gz


CMD ["/usr/bin/tail", "-f", "/tmp/foo"]
