## Running inside a Docker Container

If your existing infrastructure is container based (docker) you might be hesitant
to place remote_syslog2 on your host running wild from your container orchestration;
Or if you want to develop and test  without "tainting" your environment.

Placing remote_syslog2 inside of a base centos/ubuntu/debian image would result in a large file size (>200MB), basing it off of busybox or any small distro can leave you with a functional remote_syslog2 < 50MB.

### Prerequisite 

Install docker on your host, typically via a package manager,
be warned you may have stale packages on your system, check docker's latest documentation to insure you install the right package. (https://docs.docker.com/search/?q=install).

### Build and Run (with docker cli):

    #change version based on packages available on release page
    docker build --build-arg VERSION=v0.19 -t rs2:latest  .

After the build command a successfully built message should be produced.
`docker images` will reveal image size and name.

run (in background, _docker ps_ to confirm)

    docker run --name rs2 -d rs2:latest


If `docker ps` returns an empty table check the docker container's stdout

    docker logs rs2

This will produce errors from the container's daemonized process (remote_syslog2) which should be a decent hint as to why it crashed OR container did not start.

### Volumes and docker

Utilize a docker volume to access logs on the host or another container, between two (or more) containers share a docker volume and point the directories in `log _files.yml` at this  shared directory.

### Sending logs from the host

use the docker cli to mount a host directory OR a single file, ie (/var/log/foobar) OR /locallog.txt
    
    docker run --name rs2 -v /host/absolute/path/to/a/file/on/host/locallog.txt:/locallog.txt -d rs2:latest

confirm functionality
    
    docker logs rs2
    # output
    2017-08-06 18:40:36 INFO  remote_syslog.go:55 Connecting to logs.papertrailapp.com:514 over tls
    2017-08-06 18:40:36 INFO  remote_syslog.go:202 Forwarding file: locallog.txt
    # continual writes are "picked up by a daemon"
    lsof /host/absolute/path/to/a/file/locallog.txt
    # snippet
    1 COMMAND     PID     USER   FD   TYPE DEVICE SIZE/OFF     NODE NAME
    2 remote_sy 15465     9999    6r   REG    9,3       11 10224554 /host/absolute/path/to/a/file/locallog.txt

### Debugging

Remove the comment lines at the bottom of the dockerfile, rebuild the image, and run the container. You can use the following command `docker exec -it -u 0 sh` to "enter" the container as root, manually run remote_syslog2 via the cli and debug from there.

### Afterword

Keep the image minimal so it can be re-deployed in your enviroments. 

Use environment variables to manipulate remote_syslog2's configuration - docs.docker.com (search ENV) 
OR volume mount a configuration file `/etc/log_files.yml`

Use the docker cli for debugging/testing/development/prototyping, any other use-case should invole orchestration;

docs.docker.com (search docker-compose)

google.com (search marathon)

Managing multiple volumes (log files/directories that contain logs) is manageable and extensible with proper container orchestration; additionally steps should be taken in production environments to ensure reads/writes/truncating/rotation etc is done properly within docker volumes (fine tunning the environment)
