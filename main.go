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
	flagSet, inputOptions, outputOptions, err := parseFlags()
	if err != nil {
		return err
	}

	// Parse positional arguments
	req, err := input.ParseArgs(flagSet.Args(), os.Stdin, inputOptions)
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

func parseFlags() (*getopt.Set, *input.Options, *output.Options, error) {
	// Parse flags
	inputOptions := &input.Options{}
	outputOptions := &output.Options{}
	var ignoreStdin bool
	printFlag := "\000" // "\000" is a special value that indicates user did not specified --print

	flagSet := getopt.New()
	flagSet.SetParameters("[METHOD] URL [REQUEST_ITEM [REQUEST_ITEM ...]]")
	flagSet.BoolVarLong(&inputOptions.Form, "form", 'f', "serialize body in application/x-www-form-urlencoded")
	flagSet.StringVarLong(&printFlag, "print", 'p', "specifies what the output should contain (HBhb)")
	flagSet.BoolVarLong(&ignoreStdin, "ignore-stdin", 0, "do not attempt to read stdin")
	flagSet.Parse(os.Args)

	// Check stdin
	if !ignoreStdin && !isatty.IsTerminal(os.Stdin.Fd()) {
		inputOptions.ReadStdin = true
	}

	// Parse --print
	if err := parsePrintFlag(printFlag, outputOptions); err != nil {
		return nil, nil, nil, err
	}

	return flagSet, inputOptions, outputOptions, nil
}

func parsePrintFlag(printFlag string, outputOptions *output.Options) error {
	if printFlag == "\000" {
		// --print is not specified
		if isatty.IsTerminal(os.Stdout.Fd()) {
			outputOptions.PrintResponseHeader = true
			outputOptions.PrintResponseBody = true
		} else {
			outputOptions.PrintResponseBody = true
		}
	} else {
		for _, c := range printFlag {
			switch c {
			case 'H':
				outputOptions.PrintRequestHeader = true
			case 'B':
				outputOptions.PrintRequestBody = true
			case 'h':
				outputOptions.PrintResponseHeader = true
			case 'b':
				outputOptions.PrintResponseBody = true
			default:
				return errors.Errorf("Invalid char in --print value (must be consist of HBhb): %c", c)
			}
		}
	}
	return nil
}
