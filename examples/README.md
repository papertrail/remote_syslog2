# Examples

## com.papertrailapp.remote_syslog.plist

This is an example Mac OS X plist file.  This file should be placed at `/Library/LaunchDaemons/com.papertrailapp.remote_syslog.plist`.

## log_files.yml.example

This is a simple configuration file example.  Use it as a template for your configuration.  This file should be placed at `/etc/log_files.yml`.

## log_files.yml.example.advanced

More advanced example of above.

## remote_syslog.init.d

This is an init.d script.  Use this if your system uses init.d for startup scripts.  Place this file at `/etc/init.d/remote_syslog` and then run `chmod +x /etc/init.d/remote_syslog`.  To start the service, run `service remote_syslog start` and to run on startup, run `update-rc.d remote_syslog defaults`.

## remote_syslog.supervisor.conf

This is a supervisor configuration file.

## remote_syslog.systemd.service

This is a systemd service configuration file.  Place this file at `/etc/systemd/system/remote_syslog.service` and then run `systemctl enable remote_syslog.service` to enable the service and `systemctl start remote_syslog.service` to start it.

## remote_syslog.upstart.conf

This is an upstart configuration file.