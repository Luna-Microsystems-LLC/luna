package error

import (
	"fmt"
	"os"
)

var errors = []string {
	"no input files",
	"unexpected token",
	"expected",
	"redefinition of",
	"use of undeclared identifier",
}

var Warnings int = 0
var Errors int = 0
func Error(errno int, args string) {
	fmt.Fprintln(os.Stderr, "\033[1;39mlcc: \033[1;31merror: \033[1;39m" + errors[errno] + " " + args + "\033[0m")
	Errors = Errors + 1
	os.Exit(1)
}
