package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/nojima/httpie-go/input"
	"github.com/nojima/httpie-go/output"
	"github.com/nojima/httpie-go/request"
	"github.com/pborman/getopt"
)

func innerMain() error {
	// Parse flags
	getopt.Parse()

	// Parse positional arguments
	args := getopt.Args()
	req, err := input.ParseArgs(args)
	if err != nil {
		return err
	}

	// Send request and receive response
	resp, err := request.SendRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Print response
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()
	printer := output.NewPrettyPrinter(output.PrettyPrinterConfig{
		Writer:      writer,
		EnableColor: isatty.IsTerminal(os.Stdout.Fd()),
	})
	if err := printer.PrintStatusLine(resp); err != nil {
		return err
	}
	if err := printer.PrintHeader(resp.Header); err != nil {
		return err
	}
	writer.Flush()
	if err := printer.PrintBody(resp.Body, resp.Header.Get("Content-Type")); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := innerMain(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}
