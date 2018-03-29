package main

import (
	"testing"
	"regexp"
	"fmt"
)

func TestParseGatewayLogEntry(t *testing.T) {
	ln := `{"server":"nftgateway.qme.com","remote_addr":"139.199.43.247",` +
		`"timestamp":"2017-Oct-20 00:07:38.001","method":"HEAD","request_uri":"/",` +
		`"protocol":"HTTP/1.1","status":"200","body_bytes_sent":"0","latency":0,"response":` +
		`{"org.eclipse.jetty.server.welcome":"index.html"}}`
	rex := regexp.MustCompile(`"timestamp":".{12}(.{8}).*"latency":(\d*),"`)
	//rex := regexp.MustCompile(`"timestamp":"([^"]*)",`)
	r := rex.FindStringSubmatch(ln)
	fmt.Println(r)
	fmt.Printf("timestamp: %s\n", r[len(r)-2])
	fmt.Printf("latency: %s", r[len(r)-1])
}

func TestParseUcasLogEntry(t *testing.T) {
	ln := `2017-10-20 17:15:00,001 DEBUG [com.hundsun.jresplus.remoting.impl.CloudServiceProcessor] - ` +
		`current funtionId:306122, parsing request parameters..routeTagInfo:ctcs-gateway#1;bus_ar#1;ctcs-uc-app-as#0`
	rex := regexp.MustCompile(`.{11}(.{8}),.*current funtionId:306122, parsing request parameters`)
	r := rex.FindStringSubmatch(ln)
	fmt.Println(r)
	if len(r) > 0 {
		fmt.Printf("extracted: %+v", r[len(r)-1])
	}
}
