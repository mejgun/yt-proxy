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
