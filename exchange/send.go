package exchange

import (
	"github.com/nojima/httpie-go/input"
	"github.com/pkg/errors"
	"net/http"
)

func SendRequest(request *input.Request, options *Options) (*http.Response, error) {
	client, err := BuildHTTPClient(options)
	if err != nil {
		return nil, err
	}
	r, err := BuildHTTPRequest(request)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "sending HTTP request")
	}

	return resp, nil
}

func BuildHTTPClient(options *Options) (*http.Client, error) {
	client := http.Client{
		// Do not follow redirects
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: options.Timeout,
	}
	return &client, nil
}
