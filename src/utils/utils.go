// Package utils contains os-related utils
package utils

import (
	"fmt"
	"log"
	"os"
)

// WriteError prints error to stderr
func WriteError(err error) {
	if _, e := os.Stderr.WriteString(fmt.Sprintf("%s\n", err)); e != nil {
		log.Println("Cannot write error to stderr, error is", err)
	}
}

// WriteStdoutLn prints string to stdout
func WriteStdoutLn(str string) {
	if _, err := os.Stdout.WriteString(fmt.Sprintf("%s\n", str)); err != nil {
		log.Println("Cannot write to stdout, error is", err, "msg is", str)
	}
}
