package main

import (
	"io"
)

// ProcessOutput a struct to store process std out and std err
type ProcessOutput struct {
	Stdout io.ReadCloser
	Stderr io.ReadCloser
}
