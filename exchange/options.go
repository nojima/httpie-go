package exchange

import (
	"net/http"
	"time"
)

type Options struct {
	Timeout         time.Duration
	FollowRedirects bool
	Auth            AuthOptions
	SkipVerify      bool
	ForceHTTP1      bool
	CheckStatus     bool
	Transport       http.RoundTripper
}

type AuthOptions struct {
	Enabled  bool
	UserName string
	Password string
}
