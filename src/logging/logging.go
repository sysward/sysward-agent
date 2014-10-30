package logging

import (
	"fmt"
	"log/syslog"
	"runtime"
	"strings"
)

func LogMsg(msg string) {
	logfile, _ := syslog.New(syslog.LOG_NOTICE, "SYSWARD")
	//logfile := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	pc, _, _, _ := runtime.Caller(1)
	caller := runtime.FuncForPC(pc).Name()
	_, file, line, _ := runtime.Caller(0)
	sp := strings.Split(file, "/")
	shortPath := sp[len(sp)-2 : len(sp)]
	pathLine := fmt.Sprintf("[%s:%d]", shortPath[1], line)
	logString := fmt.Sprintf("%s{%s}:: %s", pathLine, caller, msg)
	logfile.Info(logString)
}
