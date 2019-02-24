package input

import "net/url"

type Input struct {
	Method     Method
	URL        *url.URL
	Parameters []Field
	Header     Header
	Body       Body
}

type Method string

type Header struct {
	Fields []Field
}

type BodyType int

const (
	EmptyBody BodyType = iota
	JSONBody
	FormBody
	RawBody
)

type Body struct {
	BodyType      BodyType
	Fields        []Field
	RawJSONFields []Field // used only when BodyType == JSONBody
	Files         []Field // used only when BodyType == FormBody
	Raw           []byte  // used only when BodyType == RawBody
}

type Field struct {
	Name   string
	Value  string
	IsFile bool
}
