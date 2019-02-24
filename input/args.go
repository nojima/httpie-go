package input

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/url"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

var (
	reMethod          = regexp.MustCompile(`^[a-zA-Z]+$`)
	reHeaderFieldName = regexp.MustCompile("^[-!#$%&'*+.^_|~a-zA-Z0-9]+$")
	reScheme          = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+-.]*://`)
	emptyMethod       = Method("")
)

type itemType int

const (
	unknownItem itemType = iota
	httpHeaderItem
	urlParameterItem
	dataFieldItem
	rawJSONFieldItem
	formFileFieldItem
)

type UsageError string

func (e *UsageError) Error() string {
	return string(*e)
}

func newUsageError(message string) error {
	u := UsageError(message)
	return errors.WithStack(&u)
}

func ParseArgs(args []string, stdin io.Reader, options *Options) (*Input, error) {
	var argMethod string
	var argURL string
	var argItems []string
	switch len(args) {
	case 0:
		return nil, newUsageError("URL is required")
	case 1:
		argURL = args[0]
	default:
		if reMethod.MatchString(args[0]) {
			argMethod = args[0]
			argURL = args[1]
			argItems = args[2:]
		} else {
			argURL = args[0]
			argItems = args[1:]
		}
	}

	in := Input{}

	u, err := parseURL(argURL)
	if err != nil {
		return nil, err
	}
	in.URL = u

	preferredBodyType := determinePreferredBodyType(options)
	for _, arg := range argItems {
		if err := parseItem(arg, preferredBodyType, &in); err != nil {
			return nil, err
		}
	}
	if options.ReadStdin {
		if in.Body.BodyType != EmptyBody {
			return nil, errors.New("request body (from stdin) and request item (key=value) cannot be mixed")
		}
		in.Body.BodyType = RawBody
		in.Body.Raw, err = ioutil.ReadAll(stdin)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read stdin")
		}
	}

	if argMethod != "" {
		method, err := parseMethod(argMethod)
		if err != nil {
			return nil, err
		}
		in.Method = method
	} else {
		in.Method = guessMethod(&in)
	}

	return &in, nil
}

func determinePreferredBodyType(options *Options) BodyType {
	if options.Form {
		return FormBody
	} else {
		return JSONBody
	}
}

func parseMethod(s string) (Method, error) {
	if !reMethod.MatchString(s) {
		return emptyMethod, errors.Errorf("METHOD must consist of alphabets: %s", s)
	}

	method := Method(strings.ToUpper(s))
	return method, nil
}

func guessMethod(in *Input) Method {
	if in.Body.BodyType == EmptyBody {
		return Method("GET")
	} else {
		return Method("POST")
	}
}

func parseURL(s string) (*url.URL, error) {
	defaultScheme := "http"
	defaultHost := "localhost"

	// ex) :8080/hello or /hello
	if strings.HasPrefix(s, ":") || strings.HasPrefix(s, "/") {
		s = defaultHost + s
	}

	// ex) example.com/hello
	if !reScheme.MatchString(s) {
		s = defaultScheme + "://" + s
	}

	u, err := url.Parse(s)
	if err != nil {
		return nil, newUsageError("Invalid URL: " + s)
	}
	u.Host = strings.TrimSuffix(u.Host, ":")
	if u.Path == "" {
		u.Path = "/"
	}
	return u, nil
}

func parseItem(s string, preferredBodyType BodyType, in *Input) error {
	itemType, name, value := splitItem(s)
	switch itemType {
	case dataFieldItem:
		in.Body.BodyType = preferredBodyType
		in.Body.Fields = append(in.Body.Fields, parseField(name, value))
	case rawJSONFieldItem:
		if !json.Valid([]byte(value)) {
			return errors.Errorf("invalid JSON at '%s': %s", name, value)
		}
		if preferredBodyType != JSONBody {
			return errors.New("raw JSON field item cannot be used in non-JSON body")
		}
		in.Body.BodyType = JSONBody
		in.Body.RawJSONFields = append(in.Body.RawJSONFields, parseField(name, value))
	case httpHeaderItem:
		if !isValidHeaderFieldName(name) {
			return errors.Errorf("invalid header field name: %s", name)
		}
		in.Header.Fields = append(in.Header.Fields, parseField(name, value))
	case urlParameterItem:
		in.Parameters = append(in.Parameters, parseField(name, value))
	case formFileFieldItem:
		if preferredBodyType != FormBody {
			return errors.New("form file field item cannot be used in non-form body (perhaps you meant --form?)")
		}
		in.Body.BodyType = FormBody
		in.Body.Files = append(in.Body.Files, Field{Name: name, Value: value, IsFile: true})
	default:
		return errors.Errorf("unknown request item: %s", s)
	}
	return nil
}

func splitItem(s string) (itemType, string, string) {
	for i, c := range s {
		switch c {
		case ':':
			if i+1 < len(s) && s[i+1] == '=' {
				return rawJSONFieldItem, s[:i], s[i+2:]
			} else {
				return httpHeaderItem, s[:i], s[i+1:]
			}
		case '=':
			if i+1 < len(s) && s[i+1] == '=' {
				return urlParameterItem, s[:i], s[i+2:]
			} else {
				return dataFieldItem, s[:i], s[i+1:]
			}
		case '@':
			return formFileFieldItem, s[:i], s[i+1:]
		}
	}
	return unknownItem, "", ""
}

func isValidHeaderFieldName(s string) bool {
	return reHeaderFieldName.MatchString(s)
}

func parseField(name, value string) Field {
	// TODO: handle escaped "@"
	if strings.HasPrefix(value, "@") {
		return Field{Name: name, Value: value[1:], IsFile: true}
	} else {
		return Field{Name: name, Value: value, IsFile: false}
	}
}
