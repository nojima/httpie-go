package output

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func parseURL(t *testing.T, rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		t.Fatalf("failed to parse URL: url=%s, err=%s", u, err)
	}
	return u
}

func TestPrettyPrinter_PrintStatusLine(t *testing.T) {
	// Setup
	var buffer strings.Builder
	printer := NewPrettyPrinter(PrettyPrinterConfig{
		Writer:      &buffer,
		EnableColor: false,
	})
	response := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.1",
	}

	// Exercise
	err := printer.PrintStatusLine(response.Proto, response.Status, response.StatusCode)
	if err != nil {
		t.Fatalf("unexpected error: err=%+v", err)
	}

	// Verify
	expected := "HTTP/1.1 200 OK\n"
	if buffer.String() != expected {
		t.Errorf("unexpected output: expected=%s, actual=%s", expected, buffer.String())
	}
}

func TestPrettyPrinter_PrintRequestLine(t *testing.T) {
	// Setup
	var buffer strings.Builder
	printer := NewPrettyPrinter(PrettyPrinterConfig{
		Writer:      &buffer,
		EnableColor: false,
	})
	request := &http.Request{
		Method: "GET",
		URL:    parseURL(t, "http://example.com/hello?foo=bar&hoge=piyo"),
		Proto:  "HTTP/1.1",
	}

	// Exercise
	err := printer.PrintRequestLine(request)
	if err != nil {
		t.Fatalf("unexpected error: err=%+v", err)
	}

	// Verify
	expected := "GET http://example.com/hello?foo=bar&hoge=piyo HTTP/1.1\n"
	if buffer.String() != expected {
		t.Errorf("unexpected output: expected=%s, actual=%s", expected, buffer.String())
	}
}

func TestPrettyPrinter_PrintHeader(t *testing.T) {
	// Setup
	var buffer strings.Builder
	printer := NewPrettyPrinter(PrettyPrinterConfig{
		Writer:      &buffer,
		EnableColor: false,
	})
	header := http.Header{
		"Content-Type": []string{"application/json"},
		"X-Foo":        []string{"hello", "world", "aaa"},
		"Date":         []string{"Tue, 12 Feb 2019 16:01:54 GMT"},
	}

	// Exercise
	err := printer.PrintHeader(header)
	if err != nil {
		t.Fatalf("unexpected error: err=%+v", err)
	}

	// Verify
	expected := strings.Join([]string{
		"Content-Type: application/json\n",
		"Date: Tue, 12 Feb 2019 16:01:54 GMT\n",
		"X-Foo: hello\n",
		"X-Foo: world\n",
		"X-Foo: aaa\n",
		"\n",
	}, "")
	if buffer.String() != expected {
		t.Errorf("unexpected output: expected=\n%s\n (len=%d)\nactual=\n%s\n (len=%d)",
			expected, len(expected), buffer.String(), len(buffer.String()))
	}
}

func TestPrettyPrinter_PrintBody(t *testing.T) {
	// Setup
	var buffer strings.Builder
	printer := NewPrettyPrinter(PrettyPrinterConfig{
		Writer:      &buffer,
		EnableColor: false,
	})
	body := `{"zzz": "hello \u26a1", "aaa": [3.14, true, false, "🍺"], "123": {}, "": [], "🍣": null}`

	// Exercise
	err := printer.PrintBody(strings.NewReader(body), "application/json")
	if err != nil {
		t.Fatalf("unexpected error: err=%+v", err)
	}

	// Verify
	expected := strings.Join([]string{
		`{`,
		`    "": [],`,
		`    "123": {},`,
		`    "aaa": [`,
		`        3.14,`,
		`        true,`,
		`        false,`,
		`        "🍺"`,
		`    ],`,
		`    "zzz": "hello ⚡",`, // unicode escapes should be converted to the characters they represent
		`    "🍣": null`,
		"}\n",
	}, "\n")
	if buffer.String() != expected {
		t.Errorf("unexpected output: expected=\n%s\nactual=\n%s\n", expected, buffer.String())
	}
}
