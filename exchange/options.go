package exchange

import "time"

type Options struct {
	Timeout         time.Duration
	FollowRedirects bool
	Auth            AuthOptions
	SkipVerify      bool
	ForceHTTP1      bool
}

type AuthOptions struct {
	Enabled  bool
	UserName string
	Password string
}
