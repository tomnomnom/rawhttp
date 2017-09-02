package rawhttp

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
)

type Requester interface {
	IsTLS() bool
	Host() string
	String() string
}

type RawRequest struct {
	transport string
	host      string
	request   string
}

func (r RawRequest) IsTLS() bool {
	return r.transport == "tls"
}

func (r RawRequest) Host() string {
	return r.host
}

func (r RawRequest) String() string {
	return r.request
}

type Request struct {
	TLS      bool
	Method   string
	Hostname string
	Port     string
	Path     string
	Query    string
	Fragment string
	Proto    string
	Headers  []string
	Body     string
}

func FromURL(method, rawurl string) (*Request, error) {
	r := &Request{}

	u, err := url.Parse(rawurl)
	if err != nil {
		return r, err
	}

	r.TLS = u.Scheme == "https"
	r.Method = method
	r.Hostname = u.Hostname()
	r.Port = u.Port()
	r.Path = u.Path
	r.Query = u.RawQuery
	r.Fragment = u.Fragment
	r.Proto = "HTTP/1.1"

	if r.Path == "" {
		r.Path = "/"
	}

	if r.Port == "" {
		if r.TLS {
			r.Port = "443"
		} else {
			r.Port = "80"
		}
	}

	return r, nil

}

func (r Request) IsTLS() bool {
	return r.TLS
}

func (r Request) Host() string {
	return r.Hostname + ":" + r.Port
}

func (r *Request) AddHeader(h string) {
	r.Headers = append(r.Headers, h)
}

func (r Request) fullPath() string {

	q := ""
	if r.Query != "" {
		q = "?" + r.Query
	}

	f := ""
	if r.Fragment != "" {
		f = "#" + r.Fragment
	}
	return r.Path + q + f
}

func (r Request) hasHeader(search string) bool {
	search = strings.ToLower(search) + ":"

	for _, h := range r.Headers {
		if len(h) < len(search) {
			continue
		}

		key := h[:len(search)]
		if strings.ToLower(key) == search {
			return true
		}
	}
	return false
}

func (r *Request) AutoSetHostHeader() {
	r.AddHeader(fmt.Sprintf("Host: %s", r.Hostname))
}

func (r Request) String() string {
	var b bytes.Buffer

	b.WriteString(fmt.Sprintf("%s %s %s\r\n", r.Method, r.fullPath(), r.Proto))

	for _, h := range r.Headers {
		b.WriteString(fmt.Sprintf("%s\r\n", h))
	}

	b.WriteString("\r\n")

	b.WriteString(r.Body)

	return b.String()
}

func Do(req Requester) (*Response, error) {
	var conn io.ReadWriter
	var connerr error

	// This needs timeouts because it's fairly likely
	// that something will go wrong :)
	if req.IsTLS() {
		roots, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}

		// This library is meant for doing stupid stuff, so skipping cert
		// verification is actually the right thing to do
		conf := &tls.Config{RootCAs: roots, InsecureSkipVerify: true}
		conn, connerr = tls.Dial("tcp", req.Host(), conf)

	} else {
		conn, connerr = net.Dial("tcp", req.Host())
	}

	if connerr != nil {
		return nil, connerr
	}

	fmt.Fprintf(conn, req.String())
	fmt.Fprintf(conn, "\r\n")

	return NewResponse(conn)
}
