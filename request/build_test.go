package request

import (
	"encoding/json"
	"io"
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

func isEquivalentJson(t *testing.T, json1, json2 string) bool {
	var obj1, obj2 interface{}
	if err := json.Unmarshal([]byte(json1), &obj1); err != nil {
		t.Errorf("failed to unmarshal json1: %v", err)
	}
	if err := json.Unmarshal([]byte(json2), &obj2); err != nil {
		t.Errorf("failed to unmarshal json2: %v", err)
	}
	return reflect.DeepEqual(obj1, obj2)
}

func readAll(t *testing.T, reader io.Reader) string {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Errorf("failed to read all: %s", err)
	}
	return string(b)
}

func TestBuildHttpBody_EmptyBody(t *testing.T) {
	// Setup
	fileName := makeTempFile(t, "test test")
	defer os.Remove(fileName)
	body := input.Body{
		BodyType: input.EmptyBody,
	}
	request := &input.Request{Body: body}

	// Exercise
	actual, err := buildHttpBody(request)
	if err != nil {
		t.Errorf("unexpected error: err=%+v", err)
	}

	// Verify
	expected := bodyTuple{}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("unexpected body tuple: expected=%+v, actual=%+v", expected, actual)
	}
}

func TestBuildHttpBody_JsonBody(t *testing.T) {
	// Setup
	fileName := makeTempFile(t, "test test")
	defer os.Remove(fileName)
	body := input.Body{
		BodyType: input.JsonBody,
		Fields: []input.Field{
			{Name: "foo", Value: "bar"},
			{Name: "from_file", Value: fileName, IsFile: true},
		},
		RawJsonFields: []input.Field{
			{Name: "boolean", Value: "true"},
			{Name: "array", Value: `[1, null, "hello"]`},
		},
	}
	request := &input.Request{Body: body}

	// Exercise
	bodyTuple, err := buildHttpBody(request)
	if err != nil {
		t.Errorf("unexpected error: err=%+v", err)
	}

	// Verify
	expectedBody := `{
		"foo": "bar",
		"from_file": "test test",
		"boolean": true,
		"array": [1, null, "hello"]
	}`
	actualBody := readAll(t, bodyTuple.body)
	if !isEquivalentJson(t, expectedBody, actualBody) {
		t.Errorf("unexpected body: expected=%s, actual=%s", expectedBody, actualBody)
	}
	expectedContentType := "application/json"
	if bodyTuple.contentType != expectedContentType {
		t.Errorf("unexpected content type: expected=%s, actual=%s", expectedContentType, bodyTuple.contentType)
	}
	if bodyTuple.contentLength != int64(len(actualBody)) {
		t.Errorf("invalid content length: len(body)=%v, actual=%v", len(actualBody), bodyTuple.contentLength)
	}
}

func TestBuildHttpBody_FormBody(t *testing.T) {
	// Setup
	fileName := makeTempFile(t, "love & peace")
	defer os.Remove(fileName)
	body := input.Body{
		BodyType: input.FormBody,
		Fields: []input.Field{
			{Name: "foo", Value: "bar"},
			{Name: "from_file", Value: fileName, IsFile: true},
		},
	}
	request := &input.Request{Body: body}

	// Exercise
	bodyTuple, err := buildHttpBody(request)
	if err != nil {
		t.Errorf("unexpected error: err=%+v", err)
	}

	// Verify
	expectedBody := `foo=bar&from_file=love+%26+peace`
	actualBody := readAll(t, bodyTuple.body)
	if actualBody != expectedBody {
		t.Errorf("unexpected body: expected=%s, actual=%s", expectedBody, actualBody)
	}
	expectedContentType := "application/x-www-form-urlencoded; charset=utf-8"
	if bodyTuple.contentType != expectedContentType {
		t.Errorf("unexpected content type: expected=%s, actual=%s", expectedContentType, bodyTuple.contentType)
	}
	if bodyTuple.contentLength != int64(len(actualBody)) {
		t.Errorf("invalid content length: len(body)=%v, actual=%v", len(actualBody), bodyTuple.contentLength)
	}
}
