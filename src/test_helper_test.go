package main

import "github.com/stretchr/testify/mock"

type MockRunner struct {
	mock.Mock
}

func (r *MockRunner) Run(command string, args ...string) (string, error) {
	// weird way to get around passing in just args to the helper
	pa := []string{}
	pa = append(pa, args...)
	_args := r.Mock.Called(command, pa)
	return _args.String(0), _args.Error(1)
}

type MockReader struct {
	mock.Mock
}

func (r *MockReader) ReadFile(path string) ([]byte, error) {
	args := r.Mock.Called(path)
	return args.Get(0).([]byte), args.Error(1)
}

func (r *MockReader) FileExists(path string) bool {
	args := r.Mock.Called(path)
	return args.Bool(0)
}
