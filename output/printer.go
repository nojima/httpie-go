package output

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

type Printer interface {
	PrintHeader(response *http.Response) error
	PrintBody(response *http.Response) error
}

type PlainPrinter struct {
	writer io.Writer
}

func NewPlainPrinter(writer io.Writer) Printer {
	return &PlainPrinter{
		writer: writer,
	}
}

func (p *PlainPrinter) PrintHeader(resp *http.Response) error {
	fmt.Fprintf(p.writer, "%s %s\n", resp.Proto, resp.Status)
	for name, values := range resp.Header {
		for _, value := range values {
			fmt.Fprintf(p.writer, "%s: %s\n", name, value)
		}
	}
	fmt.Fprintln(p.writer)
	return nil
}

func (p *PlainPrinter) PrintBody(resp *http.Response) error {
	_, err := io.Copy(p.writer, resp.Body)
	if err != nil {
		return errors.Wrap(err, "printing reponse body")
	}
	return nil
}

type PrettyPrinter struct {
	writer io.Writer
}

func NewPrettyPrinter(writer io.Writer) Printer {
	return &PrettyPrinter{
		writer: writer,
	}
}

func (p *PrettyPrinter) PrintHeader(resp *http.Response) error {
	fmt.Fprintf(p.writer, "%s %s\n", resp.Proto, resp.Status)

	var names []string
	for name := range resp.Header {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		values := resp.Header[name]
		for _, value := range values {
			fmt.Fprintf(p.writer, "%s: %s\n", name, value)
		}
	}

	fmt.Fprintln(p.writer)
	return nil
}

func isJSON(contentType string) bool {
	contentType = strings.TrimSpace(contentType)

	semicolon := strings.Index(contentType, ";")
	if semicolon != -1 {
		contentType = contentType[:semicolon]
	}

	return contentType == "application/json"
}

func (p *PrettyPrinter) PrintBody(resp *http.Response) error {
	if isJSON(resp.Header.Get("Content-Type")) {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "reading reponse body")
		}

		var v interface{}
		if err := json.Unmarshal(body, &v); err != nil {
			return errors.Wrap(err, "parsing response body as JSON")
		}

		encoder := json.NewEncoder(p.writer)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "    ")
		if err := encoder.Encode(v); err != nil {
			return errors.Wrap(err, "encoding JSON")
		}

		fmt.Fprintln(p.writer)
		return nil
	} else {
		_, err := io.Copy(p.writer, resp.Body)
		if err != nil {
			return errors.Wrap(err, "printing reponse body")
		}
		return nil
	}
}
