package util

import (
	"fmt"
	"os"
)

var debug bool

func init() {
	if os.Getenv("DEBUG") != "" {
		debug = true
	}
}

func Debugf(format string, a ...any) {
	if debug {
		fmt.Printf(format, a...)
	}
}

func Debug(a ...any) {
	if debug {
		fmt.Println(a...)
	}
}
