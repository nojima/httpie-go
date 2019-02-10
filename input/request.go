package input

import "net/url"

type Request struct {
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
	JsonBody
	FormBody
)

type Body struct {
	BodyType      BodyType
	Fields        []Field
	RawJsonFields []Field
}

type Field struct {
	Name   string
	Value  string
	IsFile bool
}
