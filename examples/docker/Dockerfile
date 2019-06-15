FROM alpine:3.6
MAINTAINER <gitusernameplusgithub@gmail.com>

# every system is different - every package unique
# remote_syslog2 works with docker - remote_syslog2 breaks with docker

# tested with v0.19 release - hardcoded -pull precompiled binares OR compile in dockerfile (< image)
ARG VERSION

RUN addgroup -g 9999 rs2
RUN adduser -u 9999 -D -G rs2 -k /etc/rs2 -H rs2

RUN apk update \
  && apk add ca-certificates wget \
  && update-ca-certificates 

##build arg used to pull desired versions - docker cli 
RUN wget https://github.com/papertrail/remote_syslog2/releases/download/$VERSION/remote_syslog_linux_amd64.tar.gz
RUN tar -xvf remote_syslog_linux_amd64.tar.gz

#package/tarball path changes will break dockerfile at this point
RUN mv remote_syslog/remote_syslog /usr/local/bin/remote_syslog
RUN mv /remote_syslog/example_config.yml /etc/log_files.yml

#Define the directories for -v flag
#VOLUME ["/var/log/foobar/"]
RUN touch /locallog.txt
RUN chown rs2:rs2 /locallog.txt

#run rs2 as user in production
RUN chown rs2:rs2 /usr/local/bin/remote_syslog
USER rs2
CMD ["remote_syslog", "-D" ]

## uncomment to debug container
#CMD ["/usr/bin/tail", "-f", "/tmp/foo"]
