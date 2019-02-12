package output

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type PlainPrinter struct {
	writer io.Writer
}

func NewPlainPrinter(writer io.Writer) Printer {
	return &PlainPrinter{
		writer: writer,
	}
}

func (p *PlainPrinter) PrintStatusLine(resp *http.Response) error {
	fmt.Fprintf(p.writer, "%s %s\n", resp.Proto, resp.Status)
	return nil
}

func (p *PlainPrinter) PrintHeader(header http.Header) error {
	for name, values := range header {
		for _, value := range values {
			fmt.Fprintf(p.writer, "%s: %s\n", name, value)
		}
	}
	fmt.Fprintln(p.writer)
	return nil
}

func (p *PlainPrinter) PrintBody(body io.Reader, contentType string) error {
	_, err := io.Copy(p.writer, body)
	if err != nil {
		return errors.Wrap(err, "printing response body")
	}
	return nil
}
