package cors

import "github.com/tekqer/hr"

type Options struct {
	// TODO: implement
}

func Cors(opts *Options) hr.Plugin {
	return func(h hr.Handler) hr.Handler {
		return hr.HandlerFunc(func(c *hr.Ctx) error {
			// TODO: implement
			return nil
		})
	}
}
