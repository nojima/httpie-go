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
