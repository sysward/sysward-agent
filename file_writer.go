package main

import "os"

type SystemFileWriter interface {
	AppendToFile(string, string)
}

type SyswardFileWriter struct{}

func (r SyswardFileWriter) AppendToFile(path string, contents string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(contents); err != nil {
		panic(err)
	}
}
