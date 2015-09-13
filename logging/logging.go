package logging

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"log/syslog"
	"os"
	"runtime"
	"strings"
)

func LogMsg(msg string) {
	if os.Getenv("DOCKER") == "true" {
		return
	}
	logfile, _ := syslog.New(syslog.LOG_NOTICE, "SYSWARD")
	pc, _, _, _ := runtime.Caller(1)
	caller := runtime.FuncForPC(pc).Name()
	_, file, line, _ := runtime.Caller(0)
	sp := strings.Split(file, "/")
	shortPath := sp[len(sp)-2 : len(sp)]
	pathLine := fmt.Sprintf("[%s:%d]", shortPath[1], line)
	logString := fmt.Sprintf("%s{%s}:: %s", pathLine, caller, msg)
	if os.Getenv("DEBUG") == "true" {
		log.Info(logString)
	}
	logfile.Info(logString)
}
