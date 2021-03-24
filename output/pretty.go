package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"

	"code.cloudfoundry.org/bytefmt"
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

type PrettyPrinterConfig struct {
	Writer      io.Writer
	EnableColor bool
}

type HeaderPalette struct {
	Method              aurora.Color
	URL                 aurora.Color
	Proto               aurora.Color
	SuccessfulStatus    aurora.Color
	NonSuccessfulStatus aurora.Color
	FieldName           aurora.Color
	FieldValue          aurora.Color
	FieldSeparator      aurora.Color
}

var defaultHeaderPalette = HeaderPalette{
	Method:              aurora.WhiteFg | aurora.BoldFm,
	URL:                 aurora.GreenFg | aurora.BoldFm,
	Proto:               aurora.BlueFg,
	SuccessfulStatus:    aurora.GreenFg | aurora.BoldFm,
	NonSuccessfulStatus: aurora.YellowFg | aurora.BoldFm,
	FieldName:           aurora.WhiteFg,
	FieldValue:          aurora.CyanFg,
	FieldSeparator:      aurora.WhiteFg,
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
	String:    aurora.YellowFg,
	Number:    aurora.CyanFg,
	Boolean:   aurora.RedFg | aurora.BoldFm,
	Null:      aurora.RedFg | aurora.BoldFm,
	Delimiter: aurora.WhiteFg,
}

var errMalformedJSON = errors.New("output: malformed json")

func NewPrettyPrinter(config PrettyPrinterConfig) Printer {
	return &PrettyPrinter{
		writer:        config.Writer,
		plain:         NewPlainPrinter(config.Writer),
		aurora:        aurora.NewAurora(config.EnableColor),
		headerPalette: &defaultHeaderPalette,
		jsonPalette:   &defaultJSONPalette,
		indentWidth:   4,
	}
}

func (p *PrettyPrinter) PrintStatusLine(proto string, status string, statusCode int) error {
	var statusColor aurora.Color
	if 200 <= statusCode && statusCode < 300 {
		statusColor = p.headerPalette.SuccessfulStatus
	} else {
		statusColor = p.headerPalette.NonSuccessfulStatus
	}

	fmt.Fprintf(p.writer, "%s %s\n",
		p.aurora.Colorize(proto, p.headerPalette.Proto),
		p.aurora.Colorize(status, statusColor),
	)
	return nil
}

func (p *PrettyPrinter) PrintRequestLine(req *http.Request) error {
	fmt.Fprintf(p.writer, "%s %s %s\n",
		p.aurora.Colorize(req.Method, p.headerPalette.Method),
		p.aurora.Colorize(req.URL, p.headerPalette.URL),
		p.aurora.Colorize(req.Proto, p.headerPalette.Proto),
	)
	return nil
}

func (p *PrettyPrinter) PrintHeader(header http.Header) error {
	var names []string
	for name := range header {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		values := header[name]
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

	return contentType == "application/json" || strings.HasSuffix(contentType, "+json")
}

func (p *PrettyPrinter) PrintBody(body io.Reader, contentType string) error {
	// Fallback to PlainPrinter when the body is not JSON
	if !isJSON(contentType) {
		return p.plain.PrintBody(body, contentType)
	}

	content, err := ioutil.ReadAll(body)
	if err != nil {
		return errors.Wrap(err, "reading body")
	}

	// decode JSON creating a new "token buffer" from which we will pretty-print
	// the data.
	toks, err := newTokenBuffer(json.NewDecoder(bytes.NewReader(content)))
	if err != nil || len(toks.tokens) == 0 {
		// Failed to parse body as JSON. Print as-is.
		p.writer.Write(content)
		return nil
	}

	err = p.printJSON(toks, 0)
	// errMalformedJSON errors can be ignored. This is because the JSON is
	// pre-tokenized, and therefore errMalformedJSON errors only occur when
	// the JSON ends in the middle.
	if err != nil && !errors.Is(err, errMalformedJSON) {
		return err
	}

	fmt.Fprintln(p.writer)
	return nil
}

// newTokenBuffer allows you to create a tokenBuffer which contains all the
// tokens of the given json.Decoder.
func newTokenBuffer(dec *json.Decoder) (*tokenBuffer, error) {
	dec.UseNumber()
	tks := make([]json.Token, 0, 64)
	for {
		tok, err := dec.Token()
		switch err {
		case nil:
			tks = append(tks, tok)
		case io.EOF:
			return &tokenBuffer{tokens: tks}, nil
		default:
			return nil, err
		}
	}
}

type tokenBuffer struct {
	tokens []json.Token
	pos    int
}

// endOfBody is a marker of the end of a token sequence.
type endOfBody struct{}

// token reads a new token adancing in the buffer
func (t *tokenBuffer) token() json.Token {
	if t.pos >= len(t.tokens) {
		return endOfBody{}
	}
	v := t.tokens[t.pos]
	t.pos++
	return v
}

// peek reads the next token without advancing in the buffer.
func (t *tokenBuffer) peek() json.Token {
	if t.pos >= len(t.tokens) {
		return endOfBody{}
	}
	return t.tokens[t.pos]
}

func (p *PrettyPrinter) printJSON(buf *tokenBuffer, depth int) error {
	switch v := buf.token().(type) {
	case json.Delim:
		switch v {
		case '[':
			return p.printArray(buf, depth)
		case '{':
			return p.printMap(buf, depth)
		default:
			return errors.Errorf("[BUG] wrong delim: %v", v)
		}
	case bool:
		return p.printBool(v)
	case json.Number:
		return p.printNumber(v)
	case string:
		return p.printString(v)
	case nil:
		return p.printNull()
	case endOfBody:
		return errMalformedJSON
	default:
		return errors.Errorf("[BUG] unknown value in JSON: %#v", v)
	}
}

func (p *PrettyPrinter) printNull() error {
	fmt.Fprintf(p.writer, "%s", p.aurora.Colorize("null", p.jsonPalette.Null))
	return nil
}

func (p *PrettyPrinter) printBool(v bool) error {
	var s string
	if v {
		s = "true"
	} else {
		s = "false"
	}
	fmt.Fprintf(p.writer, "%s", p.aurora.Colorize(s, p.jsonPalette.Boolean))
	return nil
}

func (p *PrettyPrinter) printNumber(n json.Number) error {
	fmt.Fprintf(p.writer, "%s", p.aurora.Colorize(n.String(), p.jsonPalette.Number))
	return nil
}

func (p *PrettyPrinter) printString(s string) error {
	b, _ := json.Marshal(s)
	fmt.Fprintf(p.writer, "%s", p.aurora.Colorize(string(b), p.jsonPalette.String))
	return nil
}

func (p *PrettyPrinter) printArray(buf *tokenBuffer, depth int) error {
	fmt.Fprintf(p.writer, "%s", p.aurora.Colorize("[", p.jsonPalette.Delimiter))

	// fast path: array is empty
	if d, ok := buf.peek().(json.Delim); ok && d == ']' {
		buf.token()
		fmt.Fprintf(p.writer, "%s", p.aurora.Colorize("]", p.jsonPalette.Delimiter))
		return nil
	}

	for {
		p.breakLine(depth + 1)

		if err := p.printJSON(buf, depth+1); err != nil {
			return err
		}

		if d, ok := buf.peek().(json.Delim); ok && d == ']' {
			// we're finished
			buf.token()
			break
		}
		fmt.Fprintf(p.writer, "%s", p.aurora.Colorize(",", p.jsonPalette.Delimiter))
	}

	p.breakLine(depth)
	fmt.Fprintf(p.writer, "%s", p.aurora.Colorize("]", p.jsonPalette.Delimiter))
	return nil
}

func (p *PrettyPrinter) printMap(buf *tokenBuffer, depth int) error {
	fmt.Fprintf(p.writer, "%s", p.aurora.Colorize("{", p.jsonPalette.Delimiter))

	// fast path: object is empty
	if d, ok := buf.peek().(json.Delim); ok && d == '}' {
		buf.token()
		fmt.Fprintf(p.writer, "%s", p.aurora.Colorize("}", p.jsonPalette.Delimiter))
		return nil
	}

	for {
		p.breakLine(depth + 1)

		key, ok := buf.token().(string)
		if !ok {
			return errMalformedJSON
		}
		encodedKey, _ := json.Marshal(key)
		fmt.Fprintf(p.writer, "%s%s ",
			p.aurora.Colorize(encodedKey, p.jsonPalette.Key),
			p.aurora.Colorize(":", p.jsonPalette.Delimiter))

		if err := p.printJSON(buf, depth+1); err != nil {
			return err
		}

		if d, ok := buf.peek().(json.Delim); ok && d == '}' {
			// we're finished
			buf.token()
			break
		}
		fmt.Fprintf(p.writer, "%s", p.aurora.Colorize(",", p.jsonPalette.Delimiter))
	}

	p.breakLine(depth)
	fmt.Fprintf(p.writer, "%s", p.aurora.Colorize("}", p.jsonPalette.Delimiter))
	return nil
}

func (p *PrettyPrinter) breakLine(depth int) {
	fmt.Fprintf(p.writer, "\n%s", strings.Repeat(" ", depth*p.indentWidth))
}

func (p *PrettyPrinter) PrintDownload(length int64, filename string) error {
	fmt.Fprintf(p.writer, "Downloading %sB to \"%s\"\n", bytefmt.ByteSize(uint64(length)), filename)
	return nil
}
