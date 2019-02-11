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
