package input

import (
	"encoding/json"
	"net/url"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

var (
	reMethod          = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	reHeaderFieldName = regexp.MustCompile("^[-!#$%&'*+.^_|~a-zA-Z0-9]+$")
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

func ParseArgs(args []string) (*Request, error) {
	request := &Request{}

	if len(args) < 1 {
		return nil, errors.New("METHOD is missing")
	}
	method, err := parseMethod(args[0])
	if err != nil {
		return nil, err
	}
	request.Method = method

	if len(args) < 2 {
		return nil, errors.New("URL is missing")
	}
	u, err := parseUrl(args[1])
	if err != nil {
		return nil, err
	}
	request.URL = u

	for _, arg := range args[2:] {
		if err := parseItem(arg, request); err != nil {
			return nil, err
		}
	}

	return request, nil
}

func parseMethod(s string) (Method, error) {
	if !reMethod.MatchString(s) {
		return emptyMethod, errors.Errorf("METHOD must consist of alphabets and numbers: %s", s)
	}
	method := Method(s)
	return method, nil
}

func parseUrl(s string) (*url.URL, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, errors.Errorf("Invalid URL: %s", s)
	}
	return u, nil
}

func parseItem(s string, request *Request) error {
	itemType, name, value := splitItem(s)
	switch itemType {
	case dataFieldItem:
		request.Body.Fields = append(request.Body.Fields, parseField(name, value))
	case rawJSONFieldItem:
		if !json.Valid([]byte(value)) {
			return errors.Errorf("invalid JSON at '%s': %s", name, value)
		}
		request.Body.RawJsonFields = append(request.Body.RawJsonFields, parseField(name, value))
	case httpHeaderItem:
		if !isValidHeaderFieldName(name) {
			return errors.Errorf("invalid header field name: %s", name)
		}
		request.Header.Fields = append(request.Header.Fields, parseField(name, value))
	case urlParameterItem:
		request.Parameters = append(request.Parameters, parseField(name, value))
	case formFileFieldItem:
		// TODO
		return errors.New("form file field item is not implemented")
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
