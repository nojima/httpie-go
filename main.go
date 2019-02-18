package httpie

import (
	"bufio"
	"os"

	"github.com/nojima/httpie-go/flags"
	"github.com/nojima/httpie-go/input"
	"github.com/nojima/httpie-go/output"
	"github.com/nojima/httpie-go/request"
	"github.com/pkg/errors"
)

func Main() error {
	// Parse flags
	args, usage, optionSet, err := flags.Parse(os.Args)
	if err != nil {
		return err
	}
	inputOptions := optionSet.InputOptions
	requestOptions := optionSet.RequestOptions
	outputOptions := optionSet.OutputOptions

	// Parse positional arguments
	req, err := input.ParseArgs(args, os.Stdin, &inputOptions)
	if _, ok := errors.Cause(err).(*input.UsageError); ok {
		usage.PrintUsage(os.Stderr)
		return err
	}
	if err != nil {
		return err
	}

	// Send request and receive response
	resp, err := request.SendRequest(req, &requestOptions)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Print response
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()
	printer := output.NewPrettyPrinter(output.PrettyPrinterConfig{
		Writer:      writer,
		EnableColor: outputOptions.EnableColor,
	})
	if outputOptions.PrintResponseHeader {
		if err := printer.PrintStatusLine(resp); err != nil {
			return err
		}
		if err := printer.PrintHeader(resp.Header); err != nil {
			return err
		}
		writer.Flush()
	}
	if outputOptions.PrintResponseBody {
		if err := printer.PrintBody(resp.Body, resp.Header.Get("Content-Type")); err != nil {
			return err
		}
	}

	return nil
}
