package exchange

import (
	"net/http"
)

func BuildHTTPClient(options *Options) (*http.Client, error) {
	checkRedirect := func(req *http.Request, via []*http.Request) error {
		// Do not follow redirects
		return http.ErrUseLastResponse
	}
	if options.FollowRedirects {
		checkRedirect = nil
	}

	client := http.Client{
		CheckRedirect: checkRedirect,
		Timeout:       options.Timeout,
	}
	return &client, nil
}
