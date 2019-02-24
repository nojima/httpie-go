package input

import (
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func mustURL(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic("Failed to parse URL: " + rawurl)
	}
	return u
}

func TestParseArgs(t *testing.T) {
	testCases := []struct {
		title         string
		args          []string
		stdin         string
		options       *Options
		expectedInput *Input
		shouldBeError bool
	}{
		{
			title: "Happy case",
			args:  []string{"GET", "http://example.com/hello"},
			expectedInput: &Input{
				Method: Method("GET"),
				URL:    mustURL("http://example.com/hello"),
			},
			shouldBeError: false,
		},
		{
			title: "Method is omitted (only host)",
			args:  []string{"localhost"},
			expectedInput: &Input{
				Method: Method("GET"),
				URL:    mustURL("http://localhost/"),
			},
		},
		{
			title: "Method is omitted (JSON body)",
			args:  []string{"example.com", "foo=bar"},
			expectedInput: &Input{
				Method: Method("POST"),
				URL:    mustURL("http://example.com/"),
				Body: Body{
					BodyType: JSONBody,
					Fields: []Field{
						{Name: "foo", Value: "bar"},
					},
				},
			},
		},
		{
			title: "Method is omitted (query parameter)",
			args:  []string{"example.com", "foo==bar"},
			expectedInput: &Input{
				Method: Method("GET"),
				URL:    mustURL("http://example.com/"),
				Parameters: []Field{
					{Name: "foo", Value: "bar"},
				},
			},
		},
		{
			title:         "URL missing",
			args:          []string{},
			expectedInput: nil,
			shouldBeError: true,
		},
		{
			title: "Lower case method",
			args:  []string{"get", "localhost"},
			expectedInput: &Input{
				Method: Method("GET"),
				URL:    mustURL("http://localhost/"),
			},
		},
		{
			title: "Read from stdin",
			args:  []string{"example.com"},
			stdin: "Hello, World!",
			options: &Options{
				ReadStdin: true,
			},
			expectedInput: &Input{
				Method: Method("POST"),
				URL:    mustURL("http://example.com/"),
				Body: Body{
					BodyType: RawBody,
					Raw:      []byte("Hello, World!"),
				},
			},
		},
		{
			title: "stdin and request items mixed",
			args:  []string{"example.com", "foo=bar"},
			stdin: "Hello, World!",
			options: &Options{
				ReadStdin: true,
			},
			shouldBeError: true,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.title, func(t *testing.T) {
			// Setup
			options := &Options{}
			if tt.options != nil {
				options = tt.options
			}

			// Exercise
			input, err := ParseArgs(tt.args, strings.NewReader(tt.stdin), options)
			if (err != nil) != tt.shouldBeError {
				t.Fatalf("unexpected error: shouldBeError=%v, err=%v", tt.shouldBeError, err)
			}
			if err != nil {
				return
			}

			// Verify
			if !reflect.DeepEqual(input, tt.expectedInput) {
				t.Errorf("unexpected input: expected=%+v, actual=%+v", tt.expectedInput, input)
			}
		})
	}
}

func TestParseItem(t *testing.T) {
	testCases := []struct {
		title                     string
		input                     string
		preferredBodyType         BodyType
		expectedBodyFields        []Field
		expectedBodyRawJSONFields []Field
		expectedBodyFiles         []Field
		expectedHeaderFields      []Field
		expectedParameters        []Field
		expectedBodyType          BodyType
		shouldBeError             bool
	}{
		{
			title:              "Data field",
			input:              "hello=world",
			expectedBodyFields: []Field{{Name: "hello", Value: "world"}},
			expectedBodyType:   JSONBody,
		},
		{
			title:              "Data field in JSON body type",
			input:              "hello=world",
			preferredBodyType:  JSONBody,
			expectedBodyFields: []Field{{Name: "hello", Value: "world"}},
			expectedBodyType:   JSONBody,
		},
		{
			title:              "Data field (form)",
			input:              "hello=world",
			preferredBodyType:  FormBody,
			expectedBodyFields: []Field{{Name: "hello", Value: "world"}},
			expectedBodyType:   FormBody,
		},
		{
			title:              "Data field with empty value",
			input:              "hello=",
			expectedBodyFields: []Field{{Name: "hello", Value: ""}},
			expectedBodyType:   JSONBody,
		},
		{
			title:              "Data field from file",
			input:              "hello=@world.txt",
			expectedBodyFields: []Field{{Name: "hello", Value: "world.txt", IsFile: true}},
			expectedBodyType:   JSONBody,
		},
		{
			title:                     "Raw JSON field",
			input:                     `hello:=[1, true, "world"]`,
			expectedBodyRawJSONFields: []Field{{Name: "hello", Value: `[1, true, "world"]`}},
			expectedBodyType:          JSONBody,
		},
		{
			title:         "Raw JSON field with invalid JSON",
			input:         `hello:={invalid: JSON}`,
			shouldBeError: true,
		},
		{
			title:             "Raw JSON field in form body type",
			input:             `hello:=[1, true, "world"]`,
			preferredBodyType: FormBody,
			shouldBeError:     true,
		},
		{
			title:                "Header field",
			input:                "X-Example:Sample Value",
			expectedHeaderFields: []Field{{Name: "X-Example", Value: "Sample Value"}},
			expectedBodyType:     EmptyBody,
		},
		{
			title:                "Header field with empty value",
			input:                "X-Example:",
			expectedHeaderFields: []Field{{Name: "X-Example", Value: ""}},
			expectedBodyType:     EmptyBody,
		},
		{
			title:         "Invalid header field name",
			input:         `Bad"header":test`,
			shouldBeError: true,
		},
		{
			title:              "URL parameter",
			input:              "hello==world",
			expectedParameters: []Field{{Name: "hello", Value: "world"}},
			expectedBodyType:   EmptyBody,
		},
		{
			title:              "URL parameter with empty value",
			input:              "hello==",
			expectedParameters: []Field{{Name: "hello", Value: ""}},
			expectedBodyType:   EmptyBody,
		},
		{
			title:             "Form file field",
			input:             "file@./hello.txt",
			preferredBodyType: FormBody,
			expectedBodyType:  FormBody,
			expectedBodyFiles: []Field{{Name: "file", Value: "./hello.txt", IsFile: true}},
		},
		{
			title:             "Form file field in JSON context",
			input:             "file@./hello.txt",
			preferredBodyType: JSONBody,
			shouldBeError:     true,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.title, func(t *testing.T) {
			// Setup
			request := Input{}
			preferredBodyType := JSONBody
			if tt.preferredBodyType != EmptyBody {
				preferredBodyType = tt.preferredBodyType
			}

			// Exercise
			err := parseItem(tt.input, &request, preferredBodyType)
			if (err != nil) != tt.shouldBeError {
				t.Fatalf("unexpected error: shouldBeError=%v, err=%v", tt.shouldBeError, err)
			}
			if err != nil {
				return
			}

			// Verify
			if !reflect.DeepEqual(request.Body.Fields, tt.expectedBodyFields) {
				t.Errorf("unexpected body field: expected=%+v, actual=%+v", tt.expectedBodyFields, request.Body.Fields)
			}
			if !reflect.DeepEqual(request.Body.RawJSONFields, tt.expectedBodyRawJSONFields) {
				t.Errorf("unexpected raw JSON body field: expected=%+v, actual=%+v", tt.expectedBodyRawJSONFields, request.Body.RawJSONFields)
			}
			if !reflect.DeepEqual(request.Body.Files, tt.expectedBodyFiles) {
				t.Errorf("unexpected files: expected=%+v, actual=%+v", tt.expectedBodyFiles, request.Body.Files)
			}
			if !reflect.DeepEqual(request.Header.Fields, tt.expectedHeaderFields) {
				t.Errorf("unexpected header field: expected=%+v, actual=%+v", tt.expectedHeaderFields, request.Header.Fields)
			}
			if !reflect.DeepEqual(request.Parameters, tt.expectedParameters) {
				t.Errorf("unexpected parameters: expected=%+v, actual=%+v", tt.expectedParameters, request.Parameters)
			}
			if request.Body.BodyType != tt.expectedBodyType {
				t.Errorf("unexpected body type: expected=%v, actual=%v", tt.expectedBodyType, request.Body.BodyType)
			}
		})
	}
}

func TestParseUrl(t *testing.T) {
	testCases := []struct {
		title    string
		input    string
		expected url.URL
	}{
		{
			title: "Typical case",
			input: "http://example.com/hello/world",
			expected: url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/hello/world",
			},
		},
		{
			title: "No scheme",
			input: "example.com/hello/world",
			expected: url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/hello/world",
			},
		},
		{
			title: "No host and port",
			input: "/hello/world",
			expected: url.URL{
				Scheme: "http",
				Host:   "localhost",
				Path:   "/hello/world",
			},
		},
		{
			title: "No host and port but has colon",
			input: ":/foo",
			expected: url.URL{
				Scheme: "http",
				Host:   "localhost",
				Path:   "/foo",
			},
		},
		{
			title: "Only colon",
			input: ":",
			expected: url.URL{
				Scheme: "http",
				Host:   "localhost",
				Path:   "/",
			},
		},
		{
			title: "No host but has port",
			input: ":8080/hello/world",
			expected: url.URL{
				Scheme: "http",
				Host:   "localhost:8080",
				Path:   "/hello/world",
			},
		},
		{
			title: "Has query parameters",
			input: "http://example.com/?q=hello&lang=ja",
			expected: url.URL{
				Scheme:   "http",
				Host:     "example.com",
				Path:     "/",
				RawQuery: "q=hello&lang=ja",
			},
		},
		{
			title: "No path",
			input: "https://example.com",
			expected: url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/",
			},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.title, func(t *testing.T) {
			// Exercise
			u, err := parseURL(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: err=%v", err)
			}

			// Verify
			if !reflect.DeepEqual(*u, tt.expected) {
				t.Errorf("unexpected result: expected=%+v, actual=%+v", tt.expected, *u)
			}
		})
	}
}
