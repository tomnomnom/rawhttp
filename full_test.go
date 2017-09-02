package rawhttp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestRaw(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Response", "check")
		fmt.Fprintln(w, "the response")
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	req := RawRequest{
		transport: "tcp",
		host:      u.Host,
		request:   "GET /anything HTTP/1.1\r\n" + "Host: localhost\r\n",
	}

	resp, err := Do(req)
	if err != nil {
		t.Errorf("want nil error, have %s", err)
	}

	have := strings.TrimSpace(string(resp.body))
	if have != "the response" {
		t.Errorf("want body to be 'the response'; have '%s'", have)
	}

	if resp.Header("Response") != "check" {
		t.Errorf("want response header to be 'check' have '%s'", resp.Header("Response"))
	}
}

func TestFromURL(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Response", "check")
		fmt.Fprintln(w, "the response")
	}))
	defer ts.Close()

	req, err := FromURL("POST", ts.URL)
	if err != nil {
		t.Fatalf("want nil error, have %s", err)
	}
	req.AutoSetHostHeader()
	req.Body = "This is some POST data"

	resp, err := Do(req)

	if err != nil {
		t.Fatalf("want nil error, have %s", err)
	}

	have := strings.TrimSpace(string(resp.body))
	if have != "the response" {
		t.Errorf("want body to be 'the response'; have '%s'", have)
	}

	if resp.Header("Response") != "check" {
		t.Errorf("want response header to be 'check' have '%s'", resp.Header("Response"))
	}
}
