package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nojima/httpie-go/input"
	"github.com/pkg/errors"
)

func SendRequest(request *input.Request) error {
	client, err := buildHttpClient()
	if err != nil {
		return err
	}
	r, err := buildHttpRequest(request)
	if err != nil {
		return err
	}

	resp, err := client.Do(r)
	if err != nil {
		return errors.Wrap(err, "sending HTTP request")
	}
	defer resp.Body.Close()

	printResponseHeader(resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "reading response body")
	}
	fmt.Printf("%s", body)

	return nil
}

func buildHttpClient() (*http.Client, error) {
	client := http.Client{
		// Do not follow redirects
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 30 * time.Second,
	}
	return &client, nil
}

func buildHttpRequest(request *input.Request) (*http.Request, error) {
	header, err := buildHttpHeader(request)
	if err != nil {
		return nil, err
	}

	body, contentType, err := buildHttpBody(request)
	if err != nil {
		return nil, err
	}

	if header.Get("Content-Type") == "" && contentType != "" {
		header.Set("Content-Type", contentType)
	}
	if header.Get("User-Agent") == "" {
		header.Set("User-Agent", "httpie-go/0.0.0")
	}

	r := http.Request{
		Method: string(request.Method),
		URL:    request.URL,
		Header: header,
		Host:   header.Get("Host"),
		Body:   body,
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

func buildHttpBody(request *input.Request) (io.ReadCloser, string, error) {
	switch request.Body.BodyType {
	case input.EmptyBody:
		return nil, "", nil

	case input.JsonBody:
		obj := map[string]interface{}{}
		for _, field := range request.Body.Fields {
			value, err := resolveFieldValue(field)
			if err != nil {
				return nil, "", err
			}
			obj[field.Name] = value
		}
		for _, field := range request.Body.RawJsonFields {
			value, err := resolveFieldValue(field)
			if err != nil {
				return nil, "", err
			}
			var v interface{}
			if err := json.Unmarshal([]byte(value), &v); err != nil {
				return nil, "", errors.Wrapf(err, "parsing JSON value of '%s'", field.Name)
			}
			obj[field.Name] = v
		}
		body, err := json.Marshal(obj)
		if err != nil {
			return nil, "", errors.Wrap(err, "marshaling JSON of HTTP body")
		}
		return ioutil.NopCloser(bytes.NewReader(body)), "application/json", nil

	case input.FormBody:
		form := url.Values{}
		for _, field := range request.Body.Fields {
			value, err := resolveFieldValue(field)
			if err != nil {
				return nil, "", err
			}
			form.Add(field.Name, value)
		}
		body := form.Encode()
		return ioutil.NopCloser(strings.NewReader(body)), "application/x-www-form-urlencoded; charset=utf-8", nil

	default:
		return nil, "", errors.Errorf("unknown body type: %v", request.Body.BodyType)
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

func printResponseHeader(resp *http.Response) {
	fmt.Printf("%s %s\n", resp.Status, resp.Proto)
	for name, values := range resp.Header {
		for _, value := range values {
			fmt.Printf("%s: %s\n", name, value)
		}
	}
	fmt.Println("")
}
