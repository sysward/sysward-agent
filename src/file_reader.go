package main

import (
	"io/ioutil"
	"os"
)

type SystemFileReader interface {
	ReadFile(string) ([]byte, error)
	FileExists(string) bool
}

type SyswardFileReader struct{}

func (r SyswardFileReader) ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func (r SyswardFileReader) FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
