package basicauth

import (
	"encoding/base64"
	"strings"

	"github.com/tekqer/hr"
)

const (
	scheme = "basic"
	realm  = "Restricted"
)

type Options struct {
	// Realm is a string describing a protected area. A realm allows a
	// server to partition up the areas it protects (if supported by a
	// scheme that allows such partitioning), and informs users about
	// which particular username/password are required. If no realm is
	// specified, clients often display a formatted hostname instead.
	Realm string
	// Validate should validates the given username and password.
	Validate func(*hr.Ctx, string, string) error
}

func BasicAuth(opts *Options) hr.Plugin {
	if len(opts.Realm) == 0 {
		opts.Realm = realm
	}
	if opts.Validate == nil {
		opts.Validate = func(*hr.Ctx, string, string) error { return nil }
	}
	return func(next hr.Handler) hr.Handler {
		return hr.HandlerFunc(func(c *hr.Ctx) error {
			auth := c.Request().Header.Get("Authorization")
			i := len(scheme)
			if len(auth) > i+1 && auth[:i] == scheme {
				b, err := base64.StdEncoding.DecodeString(auth[i+1:])
				if err != nil {
					return hr.BadRequest(err.Error())
				}
				cred := string(b)
				i = strings.IndexByte(cred, ':')
				if err := opts.Validate(c, cred[:i], cred[i+1:]); err != nil {
					return err
				}
				return next.ServeHTTP(c)
			}
			c.ResponseWriter().Header().Set("WWW-Authenticate", scheme+" realm="+opts.Realm)
			return hr.Unauthorized("")
		})
	}
}
