#!/bin/bash

### BEGIN INIT INFO
# Provides: remote_syslog
# Required-Start: $network $remote_fs $syslog
# Required-Stop: $network $remote_fs $syslog
# Default-Start: 2 3 4 5
# Default-Stop: 0 1 6
# Short-Description: Start and Stop
# Description: Runs remote_syslog
### END INIT INFO

#       /etc/init.d/remote_syslog
#
# Starts the remote_syslog daemon
#
# chkconfig: 345 90 5
# description: Runs remote_syslog
#
# processname: remote_syslog

prog="remote_syslog"
config="/etc/log_files.yml"
pid_dir="/var/run"

EXTRAOPTIONS=""

pid_file="$pid_dir/$prog.pid"

PATH=/sbin:/bin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:$PATH

RETVAL=0

is_running(){
  [ -e $pid_file ]
}

start(){
    echo -n $"Starting $prog: "

    unset HOME MAIL USER USERNAME
    $prog -c $config --pid-file=$pid_file $EXTRAOPTIONS
    RETVAL=$?
    echo
    return $RETVAL
}

stop(){
    echo -n $"Stopping $prog: "
    if (is_running); then
      kill `cat $pid_file`
      RETVAL=$?
      echo
      return $RETVAL
    else
      echo "$pid_file not found"
    fi
}

status(){
    echo -n $"Checking for $pid_file: "

    if (is_running); then
      echo "found"
    else
      echo "not found"
    fi
}

reload(){
    restart
}

restart(){
    stop
    start
}

condrestart(){
    is_running && restart
    return 0
}


# See how we were called.
case "$1" in
    start)
	start
	;;
    stop)
	stop
	;;
    status)
	status
	;;
    restart)
	restart
	;;
    reload)
	reload
	;;
    condrestart)
	condrestart
	;;
    *)
	echo $"Usage: $0 {start|stop|status|restart|condrestart|reload}"
	RETVAL=1
esac

exit $RETVAL
