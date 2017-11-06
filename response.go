package rawhttp

import (
	"bufio"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
)

// A Response wraps the HTTP response from the server
type Response struct {
	rawStatus string
	headers   []string
	body      []byte
}

// Header finds and returns the value of a header on the response.
// An empty string is returned if no match is found.
func (r Response) Header(search string) string {
	search = strings.ToLower(search)

	for _, header := range r.headers {

		p := strings.SplitN(header, ":", 2)
		if len(p) != 2 {
			continue
		}

		if strings.ToLower(p[0]) == search {
			return strings.TrimSpace(p[1])
		}
	}
	return ""
}

// ParseLocation parses the Location header of a response,
// using the initial request for context on relative URLs
func (r Response) ParseLocation(req *Request) string {
	loc := r.Header("Location")

	if loc == "" {
		return ""
	}

	// Relative locations need the context of the request
	if len(loc) > 2 && loc[:2] == "//" {
		return req.Scheme + ":" + loc
	}

	if len(loc) > 0 && loc[0] == '/' {
		return req.Scheme + "://" + req.Hostname + loc
	}

	return loc
}

// StatusLine returns the HTTP status line from the response
func (r Response) StatusLine() string {
	return r.rawStatus
}

// StatusCode returns the HTTP status code as a string; e.g. 200
func (r Response) StatusCode() string {
	parts := strings.SplitN(r.rawStatus, " ", 3)
	if len(parts) != 3 {
		return ""
	}

	return parts[1]
}

// Headers returns the response headers
func (r Response) Headers() []string {
	return r.headers
}

// Body returns the response body
func (r Response) Body() []byte {
	return r.body
}

// addHeader adds a header to the *Response
func (r *Response) addHeader(header string) {
	r.headers = append(r.headers, header)
}

// newResponse accepts an io.Reader, reads the response
// headers and body and returns a new *Response and any
// error that occured.
func newResponse(conn io.Reader) (*Response, error) {

	r := bufio.NewReader(conn)
	resp := &Response{}

	s, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	resp.rawStatus = strings.TrimSpace(s)

	for {
		line, err := r.ReadString('\n')
		line = strings.TrimSpace(line)

		if err != nil || line == "" {
			break
		}

		resp.addHeader(line)
	}

	if cl := resp.Header("Content-Length"); cl != "" {
		length, err := strconv.Atoi(cl)

		if err != nil {
			return nil, err
		}

		if length > 0 {
			b := make([]byte, length)
			_, err = io.ReadAtLeast(r, b, length)
			if err != nil {
				return nil, err
			}
			resp.body = b
		}

	} else {
		b, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}

		resp.body = b
	}

	return resp, nil
}
