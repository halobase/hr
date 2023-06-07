package cors

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/tekqer/hr"
)

var (
	defaultAllowOrigins = []string{"*"}
	defalutAllowMethods = []string{http.MethodGet, http.MethodPost, http.MethodHead}
)

type Options struct {
	// Access-Control-Allow-Origin response header indicates whether the
	// response can be shared with requesting code from the given origin.
	// See also https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Origin
	AllowOrigins []string
	// Access-Control-Allow-Methods response header specifies one or more
	// methods allowed when accessing a resource in response to a preflight
	// request.
	// See also https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Methods
	AllowMethods []string
	// Access-Control-Allow-Headers response header is used in response
	// to a preflight request which includes the Access-Control-Request-Headers
	// to indicate which HTTP headers can be used during the actual request.
	// See also https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Headers
	AllowHeaders []string
	// Access-Control-Allow-Credentials response header tells browsers whether
	// to expose the response to the frontend JavaScript code when the request's
	// credentials mode (Request.credentials) is include.
	// See also https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Credentials
	AllowCredentials bool
	// Access-Control-Max-Age response header indicates how long the results
	// of a preflight request (that is the information contained in the
	// Access-Control-Allow-Methods and Access-Control-Allow-Headers headers)
	// can be cached.
	MaxAge time.Duration
	// Access-Control-Expose-Headers response header allows a server to
	// indicate which response headers should be made available to scripts
	// running in the browser, in response to a cross-origin request. Only
	// the CORS-safelisted response headers are exposed by default. For
	// clients to be able to access other headers, the server must list them
	// using the Access-Control-Expose-Headers header.
	// CORS-safelisted response headers by now are Cache-Control, Content-Language
	// Content-Length, Content-Type, Expires, Last-Modified and Pragma.
	// See also https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Expose-Headers
	ExposeHeaders []string
	// Validate if not nil, is used to validate the Origin request header
	// regardless of whether or not it matches AllowOrigins. Otherwise,
	// AllowOrigins will be used to validate the origin.
	Validate func(*hr.Ctx, string) bool
}

func Cors(opts Options) hr.Plugin {
	if opts.AllowOrigins == nil {
		opts.AllowOrigins = defaultAllowOrigins
	}
	if opts.AllowMethods == nil {
		opts.AllowMethods = defalutAllowMethods
	}

	allowMethods := strings.Join(opts.AllowMethods, ",")
	allowHeaders := strings.Join(opts.AllowHeaders, ",")
	exposeHeaders := strings.Join(opts.ExposeHeaders, ",")
	maxAge := strconv.Itoa(int(opts.MaxAge / time.Second))

	return func(next hr.Handler) hr.Handler {
		return hr.HandlerFunc(func(c *hr.Ctx) error {
			req := c.Request()
			origin := req.Header.Get("Origin")
			preflight := req.Method == http.MethodOptions

			// This request is probably not sent from a browser, we preceed
			// executing the next handler if it is not a preflight request,
			// otherwise send a 204 (no content) response.
			if !allowOrigin(c, origin, &opts) {
				if !preflight {
					return next.ServeHTTP(c)
				}
				return c.NoContent()
			}

			header := c.ResponseWriter().Header()
			header.Add("Access-Control-Allow-Origin", origin)
			if opts.AllowCredentials {
				header.Add("Access-Control-Allow-Credentials", "true")
			}

			// The Vary header in a CORS preflight response indicates that the
			// response can vary depending on the value of the request's Origin
			// header. This means that the response may be different for differet
			// origins, and therefore the browser should not cache the response.
			// The Vary header is used to inform caches, such as intermediaries
			// or the browser cache, that the response may be different for
			// different requests with different "Origin" headers, so that they
			// will not serve a cached response that is intended for a different
			// origin. This helps ensure that the correct response is always
			// returned to the requesting origin.
			header.Add("Vary", "Origin")

			// We set the Access-Control-Expose-Headers header and proceed executing
			// the next handler if it is a cross-origin request.
			if !preflight {
				if len(exposeHeaders) > 0 {
					header.Set("Access-Control-Expose-Headers", exposeHeaders)
				}
				return next.ServeHTTP(c)
			}

			// Now we deal with the preflight response headers finally. For
			// Access-Control-Allow-Method and Access-Control-Allow-Headers,
			// if not specified by the preflight request using headers
			// Access-Control-Request-Method and Access-Control-Request-Headers,
			// we use what are given by options.
			value := req.Header.Get("Access-Control-Request-Method")
			if len(value) == 0 {
				value = allowMethods
			}
			header.Set("Access-Control-Allow-Method", value)

			value = req.Header.Get("Access-Control-Request-Headers")
			if len(value) == 0 {
				value = allowHeaders
			}
			header.Set("Access-Control-Allow-Headers", value)

			if opts.MaxAge > 0 {
				header.Set("Access-Control-Max-Age", maxAge)
			}
			return c.NoContent()
		})
	}
}

func allowOrigin(ctx *hr.Ctx, origin string, opts *Options) bool {
	if len(origin) == 0 {
		return false
	}
	if opts.Validate != nil {
		return opts.Validate(ctx, origin)
	}
	for _, v := range opts.AllowOrigins {
		if origin == v || v == "*" {
			return true
		}
	}
	return false
}
