package output

import (
	"io"
	"net/http"
)

type Printer interface {
	PrintStatusLine(response *http.Response) error
	PrintRequestLine(request *http.Request) error
	PrintHeader(header http.Header) error
	PrintBody(body io.Reader, contentType string) error
}

func NewPrinter(w io.Writer, options *Options) Printer {
	if options.EnableFormat {
		return NewPrettyPrinter(PrettyPrinterConfig{
			Writer:      w,
			EnableColor: options.EnableColor,
		})
	} else {
		return NewPlainPrinter(w)
	}
}
