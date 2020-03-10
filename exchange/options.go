package exchange

import "time"

type Options struct {
	Timeout         time.Duration
	FollowRedirects bool
	Auth            AuthOptions
	SkipVerify      bool
	ForceHTTP1      bool
	Dst             string
}

type AuthOptions struct {
	Enabled  bool
	UserName string
	Password string
}
