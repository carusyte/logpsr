package txlog

import (
	"log"
	"regexp"
	"testing"
)

func TestRegEx(t *testing.T) {
	t.Fail()
	s := `{"seridalNo":"5ab896a03c8b783da9388efb","ip":"254.128.0.0","systemCode":"TRADE","rootFuncId":"302202","callFuncId":"306737","callSerialNo":"1","currFuncId":"302202","timestamp":9490424069784039,"status":-2}`
	tag := `"serialNo":"([^"]*)"`
	rex := regexp.MustCompile(tag)
	r := rex.FindStringSubmatch(s)
	log.Printf("found:%+v", r)
}
