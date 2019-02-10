package request

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/nojima/httpie-go/input"
	"github.com/pkg/errors"
)

func buildHttpRequest(request *input.Request) (*http.Request, error) {
	header, err := buildHttpHeader(request)
	if err != nil {
		return nil, err
	}

	bodyTuple, err := buildHttpBody(request)
	if err != nil {
		return nil, err
	}

	if header.Get("Content-Type") == "" && bodyTuple.contentType != "" {
		header.Set("Content-Type", bodyTuple.contentType)
	}
	if header.Get("User-Agent") == "" {
		header.Set("User-Agent", "httpie-go/0.0.0")
	}

	r := http.Request{
		Method:        string(request.Method),
		URL:           request.URL,
		Header:        header,
		Host:          header.Get("Host"),
		Body:          bodyTuple.body,
		ContentLength: bodyTuple.contentLength,
	}
	return &r, nil
}

func buildHttpHeader(request *input.Request) (http.Header, error) {
	header := make(http.Header)
	for _, field := range request.Header.Fields {
		value, err := resolveFieldValue(field)
		if err != nil {
			return nil, err
		}
		header.Add(field.Name, value)
	}
	return header, nil
}

type bodyTuple struct {
	body          io.ReadCloser
	contentLength int64
	contentType   string
}

func buildHttpBody(request *input.Request) (bodyTuple, error) {
	switch request.Body.BodyType {
	case input.EmptyBody:
		return bodyTuple{}, nil

	case input.JsonBody:
		obj := map[string]interface{}{}
		for _, field := range request.Body.Fields {
			value, err := resolveFieldValue(field)
			if err != nil {
				return bodyTuple{}, err
			}
			obj[field.Name] = value
		}
		for _, field := range request.Body.RawJsonFields {
			value, err := resolveFieldValue(field)
			if err != nil {
				return bodyTuple{}, err
			}
			var v interface{}
			if err := json.Unmarshal([]byte(value), &v); err != nil {
				return bodyTuple{}, errors.Wrapf(err, "parsing JSON value of '%s'", field.Name)
			}
			obj[field.Name] = v
		}
		body, err := json.Marshal(obj)
		if err != nil {
			return bodyTuple{}, errors.Wrap(err, "marshaling JSON of HTTP body")
		}
		return bodyTuple{
			body:          ioutil.NopCloser(bytes.NewReader(body)),
			contentLength: int64(len(body)),
			contentType:   "application/json",
		}, nil

	case input.FormBody:
		form := url.Values{}
		for _, field := range request.Body.Fields {
			value, err := resolveFieldValue(field)
			if err != nil {
				return bodyTuple{}, err
			}
			form.Add(field.Name, value)
		}
		body := form.Encode()
		return bodyTuple{
			body:          ioutil.NopCloser(strings.NewReader(body)),
			contentLength: int64(len(body)),
			contentType:   "application/x-www-form-urlencoded; charset=utf-8",
		}, nil

	default:
		return bodyTuple{}, errors.Errorf("unknown body type: %v", request.Body.BodyType)
	}
}

func resolveFieldValue(field input.Field) (string, error) {
	if field.IsFile {
		if strings.HasPrefix(field.Value, "-") {
			// TODO
			return "", errors.New("reading field value from STDIN is not implemented")
		} else {
			data, err := ioutil.ReadFile(field.Value)
			if err != nil {
				return "", errors.Wrapf(err, "reading field value of '%s'", field.Name)
			}
			return string(data), nil
		}
	} else {
		return field.Value, nil
	}
}
