package output

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/pkg/errors"
)

type PrettyPrinter struct {
	writer        io.Writer
	plain         Printer
	aurora        aurora.Aurora
	headerPalette *HeaderPalette
}

type HeaderPalette struct {
	Proto          aurora.Color
	Status         aurora.Color
	FieldName      aurora.Color
	FieldValue     aurora.Color
	FieldSeparator aurora.Color
}

var defaultHeaderPalette = HeaderPalette{
	Proto:          aurora.BlueFg,
	Status:         aurora.BrownFg | aurora.BoldFm,
	FieldName:      aurora.GrayFg,
	FieldValue:     aurora.CyanFg,
	FieldSeparator: aurora.GrayFg,
}

type JSONPalette struct {
	Name    aurora.Color
	String  aurora.Color
	Number  aurora.Color
	Boolean aurora.Color
	Null    aurora.Color
	Symbol  aurora.Color
}

func NewPrettyPrinter(writer io.Writer) Printer {
	return &PrettyPrinter{
		writer:        writer,
		plain:         NewPlainPrinter(writer),
		aurora:        aurora.NewAurora(true),
		headerPalette: &defaultHeaderPalette,
	}
}

func (p *PrettyPrinter) PrintHeader(resp *http.Response) error {
	fmt.Fprintf(p.writer, "%s %s\n",
		p.aurora.Colorize(resp.Proto, p.headerPalette.Proto),
		p.aurora.Colorize(resp.Status, p.headerPalette.Status))

	var names []string
	for name := range resp.Header {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		values := resp.Header[name]
		for _, value := range values {
			fmt.Fprintf(p.writer, "%s%s %s\n",
				p.aurora.Colorize(name, p.headerPalette.FieldName),
				p.aurora.Colorize(":", p.headerPalette.FieldSeparator),
				p.aurora.Colorize(value, p.headerPalette.FieldValue))
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
	// Fallback to PlainPrinter when the body is not JSON
	if !isJSON(resp.Header.Get("Content-Type")) {
		return p.plain.PrintBody(resp)
	}

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
}
