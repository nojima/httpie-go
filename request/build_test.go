package request

import (
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/nojima/httpie-go/input"
)

func makeTempFile(t *testing.T, content string) string {
	tmpfile, err := ioutil.TempFile("", "httpie-go-test-")
	if err != nil {
		t.Errorf("failed to create temporary file: %v", err)
	}
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		os.Remove(tmpfile.Name())
		t.Errorf("failed to write to temporary file: %v", err)
	}
	return tmpfile.Name()
}

func TestBuildHttpHeader(t *testing.T) {
	// Setup
	fileName := makeTempFile(t, "test test")
	defer os.Remove(fileName)
	header := input.Header{
		Fields: []input.Field{
			{Name: "X-Foo", Value: "foo", IsFile: false},
			{Name: "X-From-File", Value: fileName, IsFile: true},
			{Name: "X-Multi-Value", Value: "value 1"},
			{Name: "X-Multi-Value", Value: "value 2"},
		},
	}
	request := &input.Request{Header: header}

	// Exercise
	httpHeader, err := buildHttpHeader(request)
	if err != nil {
		t.Errorf("unexpected error: err=%+v", err)
	}

	// Verify
	expected := http.Header{
		"X-Foo":         []string{"foo"},
		"X-From-File":   []string{"test test"},
		"X-Multi-Value": []string{"value 1", "value 2"},
	}
	if !reflect.DeepEqual(httpHeader, expected) {
		t.Errorf("unexpected header: expected=%v, actual=%v", expected, httpHeader)
	}
}
