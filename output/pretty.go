package output

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
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
	jsonPalette   *JSONPalette
	indentWidth   int
}

type HeaderPalette struct {
	Proto               aurora.Color
	SuccessfulStatus    aurora.Color
	NonSuccessfulStatus aurora.Color
	FieldName           aurora.Color
	FieldValue          aurora.Color
	FieldSeparator      aurora.Color
}

var defaultHeaderPalette = HeaderPalette{
	Proto:               aurora.BlueFg,
	SuccessfulStatus:    aurora.GreenFg | aurora.BoldFm,
	NonSuccessfulStatus: aurora.BrownFg | aurora.BoldFm,
	FieldName:           aurora.GrayFg,
	FieldValue:          aurora.CyanFg,
	FieldSeparator:      aurora.GrayFg,
}

type JSONPalette struct {
	Key       aurora.Color
	String    aurora.Color
	Number    aurora.Color
	Boolean   aurora.Color
	Null      aurora.Color
	Delimiter aurora.Color
}

var defaultJSONPalette = JSONPalette{
	Key:       aurora.BlueFg,
	String:    aurora.BrownFg,
	Number:    aurora.CyanFg,
	Boolean:   aurora.RedFg | aurora.BoldFm,
	Null:      aurora.RedFg | aurora.BoldFm,
	Delimiter: aurora.GrayFg,
}

func NewPrettyPrinter(writer io.Writer) Printer {
	return &PrettyPrinter{
		writer:        writer,
		plain:         NewPlainPrinter(writer),
		aurora:        aurora.NewAurora(true),
		headerPalette: &defaultHeaderPalette,
		jsonPalette:   &defaultJSONPalette,
		indentWidth:   4,
	}
}

func (p *PrettyPrinter) PrintHeader(resp *http.Response) error {
	var statusColor aurora.Color
	if 200 <= resp.StatusCode && resp.StatusCode < 300 {
		statusColor = p.headerPalette.SuccessfulStatus
	} else {
		statusColor = p.headerPalette.NonSuccessfulStatus
	}

	fmt.Fprintf(p.writer, "%s %s\n",
		p.aurora.Colorize(resp.Proto, p.headerPalette.Proto),
		p.aurora.Colorize(resp.Status, statusColor))

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
	if err := p.printJSON(v, 0); err != nil {
		return err
	}

	fmt.Fprintln(p.writer)
	return nil
}

func (p *PrettyPrinter) printJSON(v interface{}, depth int) error {
	if v == nil {
		return p.printNull()
	}
	value := reflect.ValueOf(v)
	switch value.Kind() {
	case reflect.Bool:
		return p.printBool(value)
	case reflect.Float64:
		return p.printNumber(value)
	case reflect.String:
		return p.printString(value)
	case reflect.Slice:
		return p.printArray(value, depth)
	case reflect.Map:
		return p.printMap(value, depth)
	default:
		return errors.Errorf("[BUG] unknown value in JSON: %+v", value)
	}
}

func (p *PrettyPrinter) printNull() error {
	fmt.Fprintf(p.writer, "%s", p.aurora.Colorize("null", p.jsonPalette.Null))
	return nil
}

func (p *PrettyPrinter) printBool(value reflect.Value) error {
	var s string
	if value.Bool() {
		s = "true"
	} else {
		s = "false"
	}
	fmt.Fprintf(p.writer, "%s", p.aurora.Colorize(s, p.jsonPalette.Boolean))
	return nil
}

func (p *PrettyPrinter) printNumber(value reflect.Value) error {
	fmt.Fprintf(p.writer, "%g", p.aurora.Colorize(value.Float(), p.jsonPalette.Number))
	return nil
}

func (p *PrettyPrinter) printString(value reflect.Value) error {
	s := value.String()
	b, _ := json.Marshal(s)
	fmt.Fprintf(p.writer, "%s", p.aurora.Colorize(string(b), p.jsonPalette.String))
	return nil
}

func (p *PrettyPrinter) printArray(value reflect.Value, depth int) error {
	fmt.Fprintf(p.writer, "%s", p.aurora.Colorize("[", p.jsonPalette.Delimiter))

	n := value.Len()
	for i := 0; i < n; i++ {
		p.breakLine(depth + 1)

		elem := value.Index(i)
		if err := p.printJSON(elem.Interface(), depth+1); err != nil {
			return err
		}

		if i != n-1 {
			fmt.Fprintf(p.writer, "%s", p.aurora.Colorize(",", p.jsonPalette.Delimiter))
		}
	}

	if n != 0 {
		p.breakLine(depth)
	}
	fmt.Fprintf(p.writer, "%s", p.aurora.Colorize("]", p.jsonPalette.Delimiter))
	return nil
}

func (p *PrettyPrinter) printMap(value reflect.Value, depth int) error {
	fmt.Fprintf(p.writer, "%s", p.aurora.Colorize("{", p.jsonPalette.Delimiter))

	keys := value.MapKeys()
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].String() < keys[j].String()
	})

	for i, key := range keys {
		p.breakLine(depth + 1)

		encodedKey, _ := json.Marshal(key.String())
		fmt.Fprintf(p.writer, "%s%s ",
			aurora.Colorize(encodedKey, p.jsonPalette.Key),
			aurora.Colorize(":", p.jsonPalette.Delimiter))

		elem := value.MapIndex(key)
		if err := p.printJSON(elem.Interface(), depth+1); err != nil {
			return err
		}

		if i != len(keys)-1 {
			fmt.Fprintf(p.writer, "%s", p.aurora.Colorize(",", p.jsonPalette.Delimiter))
		}
	}

	if len(keys) != 0 {
		p.breakLine(depth)
	}
	fmt.Fprintf(p.writer, "%s", p.aurora.Colorize("}", p.jsonPalette.Delimiter))
	return nil
}

func (p *PrettyPrinter) breakLine(depth int) {
	fmt.Fprintf(p.writer, "\n%s", strings.Repeat(" ", depth*p.indentWidth))
}
