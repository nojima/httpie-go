package flags

import (
	"reflect"
	"testing"
	"time"

	"github.com/nojima/httpie-go/exchange"
	"github.com/nojima/httpie-go/output"
)

func TestParse(t *testing.T) {
	args, _, optionSet, err := parse([]string{}, terminalInfo{
		stdinIsTerminal:  true,
		stdoutIsTerminal: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: err=%+v", err)
	}

	var expectedArgs []string
	if !reflect.DeepEqual(expectedArgs, args) {
		t.Errorf("unexpected returned args: expected=%v, actual=%v", expectedArgs, args)
	}
	expectedOptionSet := &OptionSet{
		ExchangeOptions: exchange.Options{
			Timeout: 30 * time.Second,
		},
		OutputOptions: output.Options{
			PrintResponseHeader: true,
			PrintResponseBody:   true,
			EnableColor:         true,
		},
	}
	if !reflect.DeepEqual(expectedOptionSet, optionSet) {
		t.Errorf("unexpected option set: expected=\n%+v\nactual=\n%+v", expectedOptionSet, optionSet)
	}
}

func TestParsePrintFlag(t *testing.T) {
	noPrintFlag := "\000"
	testCases := []struct {
		title                       string
		printFlag                   string
		verboseFlag                 bool
		headersFlag                 bool
		bodyFlag                    bool
		stdoutIsTerminal            bool
		expectedPrintRequestHeader  bool
		expectedPrintRequestBody    bool
		expectedPrintResponseHeader bool
		expectedPrintResponseBody   bool
	}{
		{
			title:                       "No flags specified (stdout is terminal)",
			printFlag:                   noPrintFlag,
			stdoutIsTerminal:            true,
			expectedPrintResponseHeader: true,
			expectedPrintResponseBody:   true,
		},
		{
			title:                     "No flags specified (stdout is NOT terminal)",
			printFlag:                 noPrintFlag,
			stdoutIsTerminal:          false,
			expectedPrintResponseBody: true,
		},
		{
			title:     `--print=""`,
			printFlag: "",
		},
		{
			title:                      `--print=H`,
			printFlag:                  "H",
			expectedPrintRequestHeader: true,
		},
		{
			title:                    `--print=B`,
			printFlag:                "B",
			expectedPrintRequestBody: true,
		},
		{
			title:                       `--print=h`,
			printFlag:                   "h",
			expectedPrintResponseHeader: true,
		},
		{
			title:                     `--print=b`,
			printFlag:                 "b",
			expectedPrintResponseBody: true,
		},
		{
			title:                       `--print=HBhb`,
			printFlag:                   "HBhb",
			expectedPrintRequestHeader:  true,
			expectedPrintRequestBody:    true,
			expectedPrintResponseHeader: true,
			expectedPrintResponseBody:   true,
		},
		{
			title:                       "--headers",
			printFlag:                   noPrintFlag,
			headersFlag:                 true,
			expectedPrintResponseHeader: true,
		},
		{
			title:                     "--body",
			printFlag:                 noPrintFlag,
			bodyFlag:                  true,
			expectedPrintResponseBody: true,
		},
		{
			title:                       "--verbose",
			printFlag:                   noPrintFlag,
			verboseFlag:                 true,
			expectedPrintRequestHeader:  true,
			expectedPrintRequestBody:    true,
			expectedPrintResponseHeader: true,
			expectedPrintResponseBody:   true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.title, func(t *testing.T) {
			options := output.Options{}
			if err := parsePrintFlag(
				tt.printFlag,
				tt.verboseFlag,
				tt.headersFlag,
				tt.bodyFlag,
				tt.stdoutIsTerminal,
				&options,
			); err != nil {
				t.Fatalf("unexpected error: err=%+v", err)
			}

			if options.PrintRequestHeader != tt.expectedPrintRequestHeader {
				t.Errorf("unexpected PrintRequestHeader: expected=%v, actual=%v",
					tt.expectedPrintRequestHeader, options.PrintRequestHeader)
			}
			if options.PrintRequestBody != tt.expectedPrintRequestBody {
				t.Errorf("unexpected PrintRequestBody: expected=%v, actual=%v",
					tt.expectedPrintRequestBody, options.PrintRequestBody)
			}
			if options.PrintResponseHeader != tt.expectedPrintResponseHeader {
				t.Errorf("unexpected PrintResponseHeader: expected=%v, actual=%v",
					tt.expectedPrintResponseHeader, options.PrintResponseHeader)
			}
			if options.PrintResponseBody != tt.expectedPrintResponseBody {
				t.Errorf("unexpected PrintResponseBody: expected=%v, actual=%v",
					tt.expectedPrintResponseBody, options.PrintResponseBody)
			}
		})
	}
}
