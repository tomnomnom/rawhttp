package main

import (
	"fmt"
	"log"

	"github.com/tomnomnom/rawhttp"
)

func main() {
	req, err := rawhttp.FromURL("POST", "https://httpbin.org")
	if err != nil {
		log.Fatal(err)
	}

	// automatically set the host header
	req.AutoSetHost()

	req.Method = "PUT"
	req.Hostname = "httpbin.org"
	req.Port = "443"
	req.Path = "/anything"
	req.Query = "one=1&two=2"
	req.Fragment = "anchor"
	req.Proto = "HTTP/1.1"
	req.EOL = "\r\n"

	req.AddHeader("Content-Type: application/x-www-form-urlencoded")

	req.Body = "username=AzureDiamond&password=hunter2"

	// automatically set the Content-Length header
	req.AutoSetContentLength()

	fmt.Printf("%s\n\n", req.String())

	resp, err := rawhttp.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("< %s\n", resp.StatusLine())
	for _, h := range resp.Headers() {
		fmt.Printf("< %s\n", h)
	}

	fmt.Printf("\n%s\n", resp.Body())
}
