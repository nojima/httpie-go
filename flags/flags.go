package flags

import (
	"io"
	"os"
	"regexp"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/nojima/httpie-go/input"
	"github.com/nojima/httpie-go/output"
	"github.com/nojima/httpie-go/request"
	"github.com/pborman/getopt"
	"github.com/pkg/errors"
)

var reNumber = regexp.MustCompile(`^[0-9.]+$`)

type FlagSet interface {
	Args() []string
	PrintUsage(w io.Writer)
}

type OptionSet struct {
	InputOptions   input.Options
	RequestOptions request.Options
	OutputOptions  output.Options
}

func Parse(args []string) (FlagSet, *OptionSet, error) {
	// Parse flags
	inputOptions := input.Options{}
	outputOptions := output.Options{}
	requestOptions := request.Options{}
	var ignoreStdin bool
	printFlag := "\000" // "\000" is a special value that indicates user did not specified --print
	timeout := "30s"

	flagSet := getopt.New()
	flagSet.SetParameters("[METHOD] URL [REQUEST_ITEM [REQUEST_ITEM ...]]")
	flagSet.BoolVarLong(&inputOptions.Form, "form", 'f', "serialize body in application/x-www-form-urlencoded")
	flagSet.StringVarLong(&printFlag, "print", 'p', "specifies what the output should contain (HBhb)")
	flagSet.BoolVarLong(&ignoreStdin, "ignore-stdin", 0, "do not attempt to read stdin")
	flagSet.StringVarLong(&timeout, "timeout", 0, "Timeout seconds that you allow the whole operation to take")
	flagSet.Parse(args)

	// Check stdin
	if !ignoreStdin && !isatty.IsTerminal(os.Stdin.Fd()) {
		inputOptions.ReadStdin = true
	}

	// Parse --print
	if err := parsePrintFlag(printFlag, &outputOptions); err != nil {
		return nil, nil, err
	}

	// Parse --timeout
	d, err := parseDurationOrSeconds(timeout)
	if err != nil {
		return nil, nil, err
	}
	requestOptions.Timeout = d

	// Color
	outputOptions.EnableColor = isatty.IsTerminal(os.Stdout.Fd())

	optionSet := &OptionSet{
		InputOptions:   inputOptions,
		RequestOptions: requestOptions,
		OutputOptions:  outputOptions,
	}
	return flagSet, optionSet, nil
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

func parseDurationOrSeconds(timeout string) (time.Duration, error) {
	if reNumber.MatchString(timeout) {
		timeout += "s"
	}
	d, err := time.ParseDuration(timeout)
	if err != nil {
		return time.Duration(0), errors.Errorf("Value of --timeout must be a number or duration string: %v", timeout)
	}
	return d, nil
}
