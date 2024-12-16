package main

import (
	"io"
	"os"
)

func cat(args []string) (err error) {
	// The spec says when there is no operands, stdin should be used.
	// See STDIN in https://pubs.opengroup.org/onlinepubs/9799919799/utilities/cat.html
	if len(args) == 0 {
		args = []string{"-"}
	}

	stdinUsed := false
	for _, arg := range args {
		if arg == "-" {
			if stdinUsed {
				// The spec says when the stdin has been used, the second occurrence of stdin is like reading from /dev/null.
				// We do not actually need to read from /dev/null.
				// See EXAMPLES in https://pubs.opengroup.org/onlinepubs/9799919799/utilities/cat.html
				continue
			}
			_, err = io.Copy(os.Stdout, os.Stdin)
			if err != nil {
				return
			}
			stdinUsed = true
		} else {
			var f *os.File
			f, err = os.Open(arg)
			if err != nil {
				return
			}
			defer f.Close()

			_, err = io.Copy(os.Stdout, f)
			if err != nil {
				return
			}
		}
	}

	return
}

func main() {
	// We do not care about the program name.
	args := os.Args[1:]
	err := cat(args)
	if err != nil {
		panic(err)
	}
}
