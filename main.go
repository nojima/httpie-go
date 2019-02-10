package main

import (
	"fmt"
	"os"

	"github.com/nojima/httpie-go/input"
	"github.com/nojima/httpie-go/request"
	"github.com/pborman/getopt"
	"github.com/pkg/errors"
)

func hello() error {
	return errors.New("Hello")
}

func main() {
	// Parse flags
	getopt.Parse()

	// Parse positional arguments
	args := getopt.Args()
	req, err := input.ParseArgs(args)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}

	if err := request.SendRequest(req); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
}
