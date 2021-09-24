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

	client := http.Client{
		CheckRedirect: checkRedirect,
		Timeout:       options.Timeout,
	}

	var transp http.RoundTripper
	if options.Transport == nil {
		transp = http.DefaultTransport.(*http.Transport).Clone()
	} else {
		transp = options.Transport
	}
	if httpTransport, ok := transp.(*http.Transport); ok {
		httpTransport.TLSClientConfig.InsecureSkipVerify = options.SkipVerify
		if options.ForceHTTP1 {
			httpTransport.TLSClientConfig.NextProtos = []string{"http/1.1", "http/1.0"}
			httpTransport.TLSNextProto = make(map[string]func(string, *tls.Conn) http.RoundTripper)
		}
	}
	client.Transport = transp

	return &client, nil
}
