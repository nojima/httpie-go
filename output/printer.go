package output

import (
	"net/http"
)

type Printer interface {
	PrintHeader(response *http.Response) error
	PrintBody(response *http.Response) error
}
