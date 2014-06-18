package main

import (
	"strings"
)

func stripPkgOutput(s string) string {
	r := strings.NewReplacer("(", "", ")", "")
	return r.Replace(s)
}
