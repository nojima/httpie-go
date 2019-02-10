package request

import (
	"net/http"
	"time"

	"github.com/nojima/httpie-go/input"
	"github.com/pkg/errors"
)

func SendRequest(request *input.Request) (*http.Response, error) {
	client, err := buildHttpClient()
	if err != nil {
		return nil, err
	}
	r, err := buildHttpRequest(request)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "sending HTTP request")
	}

	return resp, nil
}

func buildHttpClient() (*http.Client, error) {
	client := http.Client{
		// Do not follow redirects
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 30 * time.Second,
	}
	return &client, nil
}
