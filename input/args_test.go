package input

import (
	"net/url"
	"reflect"
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
		title           string
		args            []string
		expectedRequest *Request
		shouldBeError   bool
	}{
		{
			title: "Happy case",
			args:  []string{"GET", "http://example.com/hello"},
			expectedRequest: &Request{
				Method: Method("GET"),
				URL:    mustURL("http://example.com/hello"),
			},
			shouldBeError: false,
		},
		{
			title:           "Invalid method",
			args:            []string{"GET/POST", "http://example.com/hello"},
			expectedRequest: nil,
			shouldBeError:   true,
		},
		{
			title:           "Method missing",
			args:            []string{},
			expectedRequest: nil,
			shouldBeError:   true,
		},
		{
			title:           "URL missing",
			args:            []string{"POST"},
			expectedRequest: nil,
			shouldBeError:   true,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.title, func(t *testing.T) {
			request, err := ParseArgs(tt.args)
			if (err != nil) != tt.shouldBeError {
				t.Errorf("unexpected error: shouldBeError=%v, err=%v", tt.shouldBeError, err)
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(request, tt.expectedRequest) {
				t.Errorf("unexpected request: expected=%+v, actual=%+v", tt.expectedRequest, request)
			}
		})
	}
}

func TestParseItem(t *testing.T) {
	testCases := []struct {
		title                     string
		input                     string
		expectedBodyFields        []Field
		expectedBodyRawJsonFields []Field
		expectedHeaderFields      []Field
		expectedParameters        []Field
		shouldBeError             bool
	}{
		{
			title:              "Data field",
			input:              "hello=world",
			expectedBodyFields: []Field{{Name: "hello", Value: "world"}},
		},
		{
			title:              "Data field with empty value",
			input:              "hello=",
			expectedBodyFields: []Field{{Name: "hello", Value: ""}},
		},
		{
			title:              "Data field from file",
			input:              "hello=@world.txt",
			expectedBodyFields: []Field{{Name: "hello", Value: "world.txt", IsFile: true}},
		},
		{
			title:                     "Raw JSON field",
			input:                     `hello:=[1, true, "world"]`,
			expectedBodyRawJsonFields: []Field{{Name: "hello", Value: `[1, true, "world"]`}},
		},
		{
			title:         "Raw JSON field with invalid JSON",
			input:         `hello:={invalid: JSON}`,
			shouldBeError: true,
		},
		{
			title:                "Header field",
			input:                "X-Example:Sample Value",
			expectedHeaderFields: []Field{{Name: "X-Example", Value: "Sample Value"}},
		},
		{
			title:                "Header field with empty value",
			input:                "X-Example:",
			expectedHeaderFields: []Field{{Name: "X-Example", Value: ""}},
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
		},
		{
			title:              "URL parameter with empty value",
			input:              "hello==",
			expectedParameters: []Field{{Name: "hello", Value: ""}},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.title, func(t *testing.T) {
			request := Request{}
			err := parseItem(tt.input, &request)
			if (err != nil) != tt.shouldBeError {
				t.Errorf("unexpected error: shouldBeError=%v, err=%v", tt.shouldBeError, err)
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(request.Body.Fields, tt.expectedBodyFields) {
				t.Errorf("unexpected body field: expected=%+v, actual=%+v", tt.expectedBodyFields, request.Body.Fields)
			}
			if !reflect.DeepEqual(request.Body.RawJsonFields, tt.expectedBodyRawJsonFields) {
				t.Errorf("unexpected raw JSON body field: expected=%+v, actual=%+v", tt.expectedBodyRawJsonFields, tt.expectedBodyRawJsonFields)
			}
			if !reflect.DeepEqual(request.Header.Fields, tt.expectedHeaderFields) {
				t.Errorf("unexpected header field: expected=%+v, actual=%+v", tt.expectedHeaderFields, request.Header.Fields)
			}
			if !reflect.DeepEqual(request.Parameters, tt.expectedParameters) {
				t.Errorf("unexpected parameters: expected=%+v, actual=%+v", tt.expectedParameters, request.Parameters)
			}
		})
	}
}
