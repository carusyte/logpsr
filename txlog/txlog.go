package txlog

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/montanaflynn/stats"
)

type txlog struct {
	serialNo     string
	ip           string
	rootFuncID   string
	lastFuncID   string
	currFuncID   string
	systemCode   string
	callFuncID   string
	flowName     string
	callSerialNo int
	lastSerialNo int
	status       int
	ts           time.Time
}

//ParseLog parses specific logs for performance statistics.
func ParseLog() {
	//Trade:
	// fpath := []string{
	// 	`/Users/jx/Marvel/workspace/非功能测试/uat_perf_logs-1/tradeas/1/tx.2018-03-26.log`,
	// 	`/Users/jx/Marvel/workspace/非功能测试/uat_perf_logs-1/tradeas/2/tx.2018-03-26.log`,
	// 	`/Users/jx/Marvel/workspace/非功能测试/uat_perf_logs-1/tradeas/3/tx.2018-03-26.log`,
	// 	`/Users/jx/Marvel/workspace/非功能测试/uat_perf_logs-1/tradeas/4/tx.2018-03-26.log`,
	// }
	rootFuncID := "302202"

	//Delivery:
	// fpath := []string{
	// 	`/Users/jx/Marvel/workspace/非功能测试/uat_perf_logs-1/deliveryas/1/tx.2018-03-26.log`,
	// 	`/Users/jx/Marvel/workspace/非功能测试/uat_perf_logs-1/deliveryas/2/tx.2018-03-26.log`,
	// 	`/Users/jx/Marvel/workspace/非功能测试/uat_perf_logs-1/deliveryas/3/tx.2018-03-26.log`,
	// 	`/Users/jx/Marvel/workspace/非功能测试/uat_perf_logs-1/deliveryas/4/tx.2018-03-26.log`,
	// }
	// currFuncID := "302932"

	//HPS:
	// fpath := []string{
	// 	`/Users/jx/Marvel/workspace/非功能测试/uat_perf_logs-1/hpsas/1/tx.2018-03-26.log`,
	// 	`/Users/jx/Marvel/workspace/非功能测试/uat_perf_logs-1/hpsas/2/tx.2018-03-26.log`,
	// }

	//UFTS:
	fpath := []string{
		`/Users/jx/Marvel/workspace/非功能测试/uat_perf_logs-1/uftsas/1/tx.2018-03-26.log`,
		`/Users/jx/Marvel/workspace/非功能测试/uat_perf_logs-1/uftsas/2/tx.2018-03-26.log`,
	}
	parseFile(rootFuncID, fpath...)
}

func parseFile(rootFuncID string, fpath ...string) {
	var files []*os.File
	for _, f := range fpath {
		file, err := os.Open(f)
		if err != nil {
			log.Fatal(err)
			return
		}
		files = append(files, file)
	}
	defer func() {
		for _, f := range files {
			f.Close()
		}
	}()
	logs := make(map[string][]*txlog)
	for i, f := range files {
		log.Printf("scanning file: %s", fpath[i])
		rd := bufio.NewReader(f)
		count := 0
		for {
			ln, e := Readln(rd)
			if e != nil {
				if e != io.EOF {
					panic(e)
				}
				break
			}
			count++
			tx := parse(ln)
			if tx == nil {
				continue
			}
			// filter logs
			if tx.rootFuncID != rootFuncID {
				continue
			}
			logs[tx.serialNo] = append(logs[tx.serialNo], tx)
		}
		log.Printf("#lines scanned:%d", count)
	}
	getTiming(logs)
	log.Printf("#transactions parsed:%d", len(logs))
}

// Readln returns a single line (without the ending \n)
// from the input buffered reader.
// An error is returned iff there is an error with the
// buffered reader.
func Readln(r *bufio.Reader) (string, error) {
	var (
		isPrefix = true
		err      error
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return string(ln), err
}

func getTiming(logs map[string][]*txlog) {
	timing := make(map[string][]float64)
	max := .0
	var maxlog *txlog
	for _, tx := range logs {
		starts := make(map[string]time.Time)
		for i, item := range tx {
			if item.status == 0 {
				//busi log (dao method)
				fid := funcID(item)
				delay := item.ts.Sub(tx[i-1].ts).Seconds()
				timing[fid] = append(timing[fid], delay)
				if delay > max {
					maxlog = item
					max = delay
				}
			} else if item.status < 0 && item.status != -9 {
				starts[funcID(item)] = item.ts
			} else if item.status > 0 {
				funcID := funcID(item)
				delay := item.ts.Sub(starts[funcID]).Seconds()
				timing[funcID] = append(timing[funcID], delay)
				if delay > max {
					maxlog = item
					max = delay
				}
			}
		}
	}
	var ids fids
	tmap := make(map[string]float64)
	for id, delays := range timing {
		ids = append(ids, id)
		m, e := stats.Mean(delays)
		if e != nil {
			log.Panicf("failed to calculate mean for %s", id)
		}
		tmap[id] = m
	}
	sort.Sort(ids)
	for _, id := range ids {
		fmt.Printf("%s: %.3f\n", id, tmap[id])
	}
	fmt.Printf("Max Duration: %.3f\n", max)
	fmt.Printf("Max TxLog: %+v\n", maxlog)
}

type fids []string

func (s fids) Len() int      { return len(s) }
func (s fids) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s fids) Less(i, j int) bool {
	a, b := s[i], s[j]
	sa := strings.Split(a, "_")
	ia, e := strconv.Atoi(sa[0])
	if e != nil {
		panic(e)
	}
	sb := strings.Split(b, "_")
	ib, e := strconv.Atoi(sb[0])
	if e != nil {
		panic(e)
	}
	if ia != ib {
		return ia < ib
	}
	return sa[1] < sb[1]
}

func funcID(item *txlog) string {
	if item.callSerialNo != 0 {
		return fmt.Sprintf("%d_%s", item.callSerialNo, item.callFuncID)
	}
	if item.flowName != "" {
		return fmt.Sprintf("%d_%s", item.lastSerialNo, item.flowName)
	}
	return "0_" + item.currFuncID
}

func parse(ln string) *txlog {
	ts, data := ln[:23], ln[24:]
	t, e := time.Parse(`2006-01-02 15:04:05.000`, strings.Replace(ts, ",", ".", 1))
	if e != nil {
		log.Println(e)
		return nil
	}
	tx := new(txlog)
	tx.serialNo = extract(data, `"serialNo":"([^"]*)"`)
	tx.ip = extract(data, `"ip":"([^"]*)"`)
	tx.rootFuncID = extract(data, `"rootFuncId":"([^"]*)"`)
	tx.lastFuncID = extract(data, `"lastFuncId":"([^"]*)"`)
	tx.currFuncID = extract(data, `"currFuncId":"([^"]*)"`)
	tx.systemCode = extract(data, `"systemCode":"([^"]*)"`)
	tx.callFuncID = extract(data, `"callFuncId":"([^"]*)"`)
	tx.flowName = extract(data, `"flowName":"([^"]*)"`)
	val := extract(data, `"callSerialNo":"([^"]*)"`)
	if val != "" {
		csn, e := strconv.Atoi(val)
		if e != nil {
			log.Printf("unable to parse callSerialNo:%+v\n%+v", e, data)
		}
		tx.callSerialNo = csn
	}
	val = extract(data, `"lastSerialNo":"([^"]*)"`)
	if val != "" {
		lsn, e := strconv.Atoi(val)
		if e != nil {
			log.Printf("unable to parse lastSerialNo:%+v\n%+v", e, data)
		}
		tx.lastSerialNo = lsn
	}
	status, e := strconv.Atoi(extract(data, `"status":([^"]*)[,\}]`))
	if e != nil {
		log.Printf("unable to parse status:%+v\n%+v", e, data)
	}
	tx.status = status
	tx.ts = t
	return tx
}

func extract(str, rx string) string {
	rex := regexp.MustCompile(rx)
	r := rex.FindStringSubmatch(str)
	if len(r) > 0 {
		return r[len(r)-1]
	}
	return ""
}
