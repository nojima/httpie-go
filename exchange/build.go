package exchange

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/nojima/httpie-go/input"
	"github.com/nojima/httpie-go/version"
	"github.com/pkg/errors"
)

func BuildHTTPRequest(in *input.Input) (*http.Request, error) {
	u, err := buildURL(in)
	if err != nil {
		return nil, err
	}

	header, err := buildHTTPHeader(in)
	if err != nil {
		return nil, err
	}

	bodyTuple, err := buildHTTPBody(in)
	if err != nil {
		return nil, err
	}

	if header.Get("Content-Type") == "" && bodyTuple.contentType != "" {
		header.Set("Content-Type", bodyTuple.contentType)
	}
	if header.Get("User-Agent") == "" {
		header.Set("User-Agent", fmt.Sprintf("httpie-go/%s", version.Current()))
	}

	r := http.Request{
		Method:        string(in.Method),
		URL:           u,
		Header:        header,
		Host:          header.Get("Host"),
		Body:          bodyTuple.body,
		ContentLength: bodyTuple.contentLength,
	}
	return &r, nil
}

func buildURL(in *input.Input) (*url.URL, error) {
	q, err := url.ParseQuery(in.URL.RawQuery)
	if err != nil {
		return nil, errors.Wrap(err, "parsing query string")
	}
	for _, field := range in.Parameters {
		value, err := resolveFieldValue(field)
		if err != nil {
			return nil, err
		}
		q.Add(field.Name, value)
	}

	u := *in.URL
	u.RawQuery = q.Encode()
	return &u, nil
}

func buildHTTPHeader(in *input.Input) (http.Header, error) {
	header := make(http.Header)
	for _, field := range in.Header.Fields {
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

func buildHTTPBody(in *input.Input) (bodyTuple, error) {
	switch in.Body.BodyType {
	case input.EmptyBody:
		return bodyTuple{}, nil
	case input.JSONBody:
		return buildJSONBody(in)
	case input.FormBody:
		return buildFormBody(in)
	case input.RawBody:
		return buildRawBody(in)
	default:
		return bodyTuple{}, errors.Errorf("unknown body type: %v", in.Body.BodyType)
	}
}

func buildJSONBody(in *input.Input) (bodyTuple, error) {
	obj := map[string]interface{}{}
	for _, field := range in.Body.Fields {
		value, err := resolveFieldValue(field)
		if err != nil {
			return bodyTuple{}, err
		}
		obj[field.Name] = value
	}
	for _, field := range in.Body.RawJSONFields {
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
}

func buildFormBody(in *input.Input) (bodyTuple, error) {
	if len(in.Body.Files) > 0 {
		return buildMultipartBody(in)
	} else {
		return buildURLEncodedBody(in)
	}
}

func buildURLEncodedBody(in *input.Input) (bodyTuple, error) {
	form := url.Values{}
	for _, field := range in.Body.Fields {
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
}

func buildMultipartBody(in *input.Input) (bodyTuple, error) {
	var buffer bytes.Buffer
	multipartWriter := multipart.NewWriter(&buffer)

	for _, field := range in.Body.Fields {
		if err := buildInlinePart(field, multipartWriter); err != nil {
			return bodyTuple{}, err
		}
	}
	for _, field := range in.Body.Files {
		if err := buildFilePart(field, multipartWriter); err != nil {
			return bodyTuple{}, err
		}
	}

	multipartWriter.Close()

	body := buffer.Bytes()
	return bodyTuple{
		body:          ioutil.NopCloser(bytes.NewReader(body)),
		contentLength: int64(len(body)),
		contentType:   multipartWriter.FormDataContentType(),
	}, nil
}

func buildInlinePart(field input.Field, multipartWriter *multipart.Writer) error {
	value, err := resolveFieldValue(field)
	if err != nil {
		return err
	}

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", buildContentDisposition(field.Name, ""))
	w, err := multipartWriter.CreatePart(h)
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(value)); err != nil {
		return err
	}
	return nil
}

func buildFilePart(field input.Field, multipartWriter *multipart.Writer) error {
	h := make(textproto.MIMEHeader)
	filename := path.Base(field.Value)
	h.Set("Content-Disposition", buildContentDisposition(field.Name, filename))
	w, err := multipartWriter.CreatePart(h)
	if err != nil {
		return err
	}

	file, err := os.Open(field.Value)
	if err != nil {
		return errors.Wrapf(err, "failed to open '%s'", field.Value)
	}
	defer file.Close()

	if _, err := io.Copy(w, file); err != nil {
		return errors.Wrapf(err, "failed to read from '%s'", field.Value)
	}
	return nil
}

func buildContentDisposition(name string, filename string) string {
	var buffer bytes.Buffer
	buffer.WriteString("form-data")

	if name != "" {
		if needEscape(name) {
			fmt.Fprintf(&buffer, `; name*=utf-8''%s`, url.PathEscape(name))
		} else {
			fmt.Fprintf(&buffer, `; name="%s"`, name)
		}
	}

	if filename != "" {
		if needEscape(filename) {
			fmt.Fprintf(&buffer, `; filename*=utf-8''%s`, url.PathEscape(filename))
		} else {
			fmt.Fprintf(&buffer, `; filename="%s"`, filename)
		}
	}

	return buffer.String()
}

func needEscape(s string) bool {
	for _, c := range s {
		if c > 127 {
			return true
		}
		if c < 32 && c != '\t' {
			return true
		}
		if c == '"' || c == '\\' {
			return true
		}
	}
	return false
}

func buildRawBody(in *input.Input) (bodyTuple, error) {
	return bodyTuple{
		body:          ioutil.NopCloser(bytes.NewReader(in.Body.Raw)),
		contentLength: int64(len(in.Body.Raw)),
		contentType:   "application/json",
	}, nil
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
