## Running inside a Docker Container

If your existing infrastructure is container based (docker) you might be hesitant
to place remote_syslog2 on your host running wild from your container orchestration;
Or if you want to develop and test  without "tainting" your enviroment.

Placing remote_syslog2 inside of a base centos/ubuntu/debian would result in a large (docker) image, basing it off of busybox or any small distro (tinycore anyone?) can leave you with a functional remote_syslog2 < 50MB.

### Prerequiste 

Install docker on your host, typically via a package manager,
be warned you may have stale packages on your system, check docker's latest documentation to insure you install the right package. (https://docs.docker.com/search/?q=install).

### Build and Run (with docker cli):

    # same directory as Dockerfile
    # change version based on packages available on release page
    docker build --build-arg VERSION=v0.19 -t rs2  .

A successful build message should be produced after the build command.
`docker images` will reveal image size and name.

run (in background, _docker ps_ to confirm)

    docker run --name rs2 -d rs2

If `docker ps` returns an empty table check the docker container's stdout

    docker logs rs2

This will produce errors from the container's daemonized process (remote_syslog2) which should be a decent hint as to why it crashed.

###Volumes and docker

Utilize a docker volume to access logs on the host or another container, between two (or more) containers share a docker volume and point the directories in `log _files.yml` at this directory

