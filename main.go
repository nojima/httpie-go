package httpie

import (
	"bufio"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/nojima/httpie-go/input"
	"github.com/nojima/httpie-go/output"
	"github.com/nojima/httpie-go/request"
	"github.com/pborman/getopt"
	"github.com/pkg/errors"
)

func Main() error {
	// Parse flags
	options := &input.Options{}
	var ignoreStdin bool
	flagSet := getopt.New()
	flagSet.SetParameters("[METHOD] URL [REQUEST_ITEM [REQUEST_ITEM ...]]")
	flagSet.BoolVarLong(&options.Form, "form", 'f', "serialize body in application/x-www-form-urlencoded")
	flagSet.BoolVarLong(&ignoreStdin, "ignore-stdin", 0, "do not attempt to read stdin")
	flagSet.Parse(os.Args)

	// Check stdin
	if !ignoreStdin && !isatty.IsTerminal(os.Stdin.Fd()) {
		options.ReadStdin = true
	}

	// Parse positional arguments
	req, err := input.ParseArgs(flagSet.Args(), os.Stdin, options)
	if _, ok := errors.Cause(err).(*input.UsageError); ok {
		flagSet.PrintUsage(os.Stderr)
		return err
	}
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
