package main

import (
	"testing"
	"regexp"
	"fmt"
)

func TestParseGatewayLogEntry(t *testing.T) {
	ln := `{"server":"nftgateway.qme.com","remote_addr":"139.199.43.247","timestamp":"2017-Oct-20 00:07:38.001","method":"HEAD","request_uri":"/","protocol":"HTTP/1.1","status":"200","body_bytes_sent":"0","latency":0,"response":{"org.eclipse.jetty.server.welcome":"index.html"}}`
	rex := regexp.MustCompile(`"timestamp":".{12}(.{8}).{4}",`)
	//rex := regexp.MustCompile(`"timestamp":"([^"]*)",`)
	r := rex.FindStringSubmatch(ln)
	fmt.Println(r)
}
