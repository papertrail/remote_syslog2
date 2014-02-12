package utils

const CanDaemonize = false

func Daemonize(logFilePath, pidFilePath string) {
	panic("cannot daemonize on windows")
}
