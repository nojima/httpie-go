package exchange

import (
	"net/http"
)

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
