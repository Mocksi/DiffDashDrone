//go:build mage

package main

import (
	"github.com/magefile/mage/sh"
)

// Build go mod download and then build the binary
func Build() error {
	if err := sh.Run("go", "mod", "download"); err != nil {
		return err
	}
	return sh.Run("go", "build", "-o", "bin/diffdash", "main.go")
}

// Test run tests
func Test() error {
	return sh.Run("go", "test", "-v", "./...")
}

// Lint run linter
func Lint() error {
	return sh.Run("go", "fmt")
}

// Run execute code via go run
func Run() error {
	return sh.Run("go", "run", "main.go")
}
