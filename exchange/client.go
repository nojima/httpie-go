package exchange

import (
	"crypto/tls"
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

	transp := http.DefaultTransport.(*http.Transport).Clone()
	transp.TLSClientConfig.InsecureSkipVerify = options.SkipVerify
	if options.ForceHTTP1 {
		transp.TLSClientConfig.NextProtos = []string{"http/1.1", "http/1.0"}
		transp.TLSNextProto = make(map[string]func(string, *tls.Conn) http.RoundTripper)
	}

	client := http.Client{
		CheckRedirect: checkRedirect,
		Timeout:       options.Timeout,
		Transport:     transp,
	}

	return &client, nil
}
