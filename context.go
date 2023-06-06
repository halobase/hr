package hr

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"mime/multipart"
	"net/http"
	"net/url"
)

type Ctx struct {
	context.Context
	req   *http.Request
	rw    http.ResponseWriter
	query url.Values
	vars  Vars

	wroteHeader bool
}

// Request returns the underlying HTTP request.
func (c Ctx) Request() *http.Request {
	return c.req
}

// ResponseWriter returns the underlying response writer.
func (c Ctx) ResponseWriter() http.ResponseWriter {
	return c.rw
}

// Vars returns the named route variables
func (c Ctx) Vars() Vars {
	return c.vars
}

// Query parses the URL query string and returns the corresponding
// values. It silently discards malformed value pairs.
func (c Ctx) Query() url.Values {
	if c.query == nil {
		c.query = c.req.URL.Query()
	}
	return c.query
}

// Form returns the parsed form data including both the URL field's
// query parameters and the PATCH, POST, or PUT form data.
func (c Ctx) Form() (url.Values, error) {
	// we don't bother to check if the form is parsed or not
	// because ParseForm will do it for us.
	if err := c.req.ParseForm(); err != nil {
		return nil, err
	}
	return c.req.Form, nil
}

// MultipartForm returns parsed multipart form, including file uploads.
func (c Ctx) MultipartForm() (*multipart.Form, error) {
	// we don't bother to check if the form is parsed or not
	// because ParseMultipartForm will do it for us.
	if err := c.req.ParseMultipartForm(32 << 20); err != nil { // 32MiB
		return nil, err
	}
	return c.req.MultipartForm, nil
}

// Bind deserializes data from the request to a Go struct. Where data
// is extracted is specified by Content-Type. The URL query string if
// presented is also deserialized to the struct if there are fields
// tagged with `query`. Fields will be validated if they are tagged with
// `validate`. Any error occurried during the call will be returned
// after wrapped with an hr.Error that results in a response with a
// 404 (bad request) status code. NOTE that v must be a pointer.
//
// Example:
//
//	type User struct {
//	    Name string `form:"name"`
//	    Age  int    `form:"age"`
//	}
//
//	func createUser(c hr.Ctx) error {
//	    var u User
//	    if err := c.Bind(&u); err != nil {
//	        return err
//	    }
//	    // do something with u
//	}
func (c Ctx) Bind(v interface{}) error {
	req := c.req

	if err := c.BindPath(v); err != nil {
		return err
	}

	switch req.Method {
	case http.MethodGet, http.MethodDelete, http.MethodHead:
		if err := bind(v, req.URL.Query(), "query"); err != nil {
			return BadRequest(err.Error())
		}
		return nil
	}

	ctype := req.Header.Get("Content-Type")
	switch ctype {
	case "application/json":
		defer req.Body.Close()
		if err := json.NewDecoder(req.Body).Decode(v); err != nil {
			return BadRequest(err.Error())
		}
	case "application/xml":
		defer req.Body.Close()
		if err := xml.NewDecoder(req.Body).Decode(v); err != nil {
			return BadRequest(err.Error())
		}
	case "application/gob":
		defer req.Body.Close()
		if err := gob.NewDecoder(req.Body).Decode(v); err != nil {
			return BadRequest(err.Error())
		}
	case "application/x-www-form-urlencoded":
		if err := req.ParseForm(); err != nil {
			return BadRequest(err.Error())
		}
		if err := bind(v, req.Form, "form"); err != nil {
			return BadRequest(err.Error())
		}
	default:
		return BadRequest("bad content type")
	}
	return nil
}

func (c Ctx) BindHeader(v interface{}) error {
	if err := bind(v, c.req.Header, "header"); err != nil {
		return BadRequest(err.Error())
	}
	return nil
}

func (c Ctx) BindPath(v interface{}) error {
	vars := c.vars
	if len(vars) == 0 {
		return nil
	}
	vals := make(map[string][]string, len(vars))
	for _, v := range vars {
		vals[v.Key] = []string{v.Value}
	}
	if err := bind(v, vals, "path"); err != nil {
		return BadRequest(err.Error())
	}
	return nil
}

// WriteHeader sends an HTTP response header with the provided status
// code and should not be called more than once. Invocation after the
// first one takes no effect.
func (c Ctx) WriteHeader(code int) {
	if !c.wroteHeader {
		c.rw.WriteHeader(code)
	}
}

func (c Ctx) Write(p []byte) (int, error) {
	return c.rw.Write(p)
}

func (c Ctx) JSON(v interface{}) error {
	w := c.rw
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(v)
}

func (c Ctx) XML(v interface{}) error {
	w := c.rw
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	return xml.NewEncoder(w).Encode(v)
}

func (c Ctx) Gob(v interface{}) error {
	w := c.rw
	w.Header().Set("Content-Type", "application/gob")
	w.WriteHeader(http.StatusOK)
	return gob.NewEncoder(w).Encode(v)
}
