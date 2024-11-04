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
	logfile, err := syslog.New(syslog.LOG_NOTICE, "SYSWARD")
	if err != nil {
		fmt.Println("Error writing to syslog: ", err)
		return
	}
	defer logfile.Close()
	if os.Getenv("LOG_FILE") != "" {
		f, err := os.OpenFile(os.Getenv("LOG_FILE"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("Error opening log file: ", err)
			return
		}
		defer f.Close()
		log.SetOutput(f)
	} else {
		log.SetOutput(logfile)
	}
	pc, _, _, _ := runtime.Caller(1)
	caller := runtime.FuncForPC(pc).Name()
	_, file, line, _ := runtime.Caller(0)
	sp := strings.Split(file, "/")
	shortPath := sp[len(sp)-2 : len(sp)]
	pathLine := fmt.Sprintf("[%s:%d]", shortPath[1], line)
	logString := fmt.Sprintf("%s{%s}:: %s", pathLine, caller, msg)
	if os.Getenv("DEBUG") == "true" {
		log.SetOutput(os.Stdout)
		log.Info(logString)
	}
	log.Info(logString)
}
