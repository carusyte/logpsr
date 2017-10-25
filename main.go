package main

import (
	"bufio"
	"os"
	"log"
	"strings"
	"regexp"
	"sort"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func main() {
	f := "/Users/jx/Downloads/ctcs-gateway.log.2017-10-23"
	tc, e := ParseFile(f)
	if e != nil {
		log.Panicln(e)
	}
	tc.Chop("14:00:00", "18:18:00")
	graph("points_1023.png", tc)
}

func graph(dstFile string, tcs ... *TsCounter) {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "TPS Graph"
	p.X.Label.Text = "Time"
	p.Y.Label.Text = "TPS"

	for _, tc := range tcs {
		tms, tps := tc.Timestamps, tc.Counts
		pts := make(plotter.XYs, len(tms))
		for i := range pts {
			pts[i].X = float64(i)
			pts[i].Y = float64(tps[i])
		}
		xlbs := make([]string, len(tms))
		gap := len(tms) / 7
		for i := range xlbs {
			if i%gap == 0 {
				xlbs[i] = tms[i]
			} else {
				xlbs[i] = ""
			}
		}
		p.NominalX(xlbs...)

		err = plotutil.AddLinePoints(p,
			tc.Name, pts)
		if err != nil {
			panic(err)
		}
	}

	// Save the plot to a PNG file.
	if err := p.Save(20*vg.Inch, 10*vg.Inch, dstFile); err != nil {
		panic(err)
	}
}

func ParseFile(filePath string) (tc *TsCounter, e error) {
	lns, e := ParseLines(filePath)
	if e != nil {
		return nil, e
	}
	var ftype string
	if strings.HasPrefix(filePath, "ctcs-gateway.log") {
		ftype = "gateway"
	}
	log.Printf("%s lines: %d", ftype, len(lns))
	tc = new(TsCounter)
	tc.Name = ftype
	//{"server":"nftgateway.qme.com","remote_addr":"139.199.43.247",
	// "timestamp":"2017-Oct-20 00:00:13.000",
	// "method":"HEAD","request_uri":"/","protocol":"HTTP/1.1","status":"200","body_bytes_sent":"0","latency":1,"response":{"org.eclipse.jetty.server.welcome":"index.html"}}
	rex := regexp.MustCompile(`"timestamp":".{12}(.{8}).{4}",`)
	for _, ln := range lns {
		r := rex.FindStringSubmatch(ln)
		if len(r) > 0 {
			tc.Add(r[len(r)-1])
		}
	}
	//ts, cs := tc.OrderByTime()
	//for i, _ := range ts {
	//	fmt.Printf("%s\t%d", ts[i], cs[i])
	//	fmt.Println()
	//}
	return tc, nil
}

func ParseLines(filePath string) ([]string, error) {
	inputFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer inputFile.Close()

	scanner := bufio.NewScanner(inputFile)
	var results []string
	for scanner.Scan() {
		results = append(results, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

type TsCounter struct {
	Name       string
	Timestamps []string
	Counter    map[string]int
	Counts     []int
	ordered    bool
}

func (tc *TsCounter) Add(timeStamp string) {
	if tc.Counter == nil {
		tc.Timestamps = make([]string, 0, 16)
		tc.Counter = make(map[string]int)
	}
	if c, ok := tc.Counter[timeStamp]; ok {
		tc.Counter[timeStamp] = c + 1
	} else {
		tc.Counter[timeStamp] = 1
		tc.Timestamps = append(tc.Timestamps, timeStamp)
	}
}

func (tc *TsCounter) Chop(start, end string) {
	inclSt, inclEd := true, true
	if strings.HasPrefix(start, "(") {
		inclSt = false
		start = start[1:]
	}
	if strings.HasSuffix(end, ")") {
		inclEd = false
		end = end[:len(end)-1]
	}

	if !tc.ordered {
		tc.OrderByTime()
	}
	ts, cs := tc.Timestamps, tc.Counts
	sidx, eidx := 0, len(ts)
	for i, t := range ts {
		if sidx == 0 && start != "" {
			if inclSt {
				if start <= t {
					sidx = i
				}
			} else {
				if start < t {
					sidx = i
				}
			}
		}
		if end != "" {
			if inclEd {
				if end == t {
					eidx = i
					break
				} else if end < t {
					eidx = i - 1
					break
				}
			} else {
				if end <= t {
					eidx = i - 1
					break
				}
			}
		} else if sidx != 0 || start == "" {
			break
		}
	}
	tc.Timestamps = ts[sidx:eidx]
	tc.Counts = cs[sidx:eidx]
}

func (tc *TsCounter) OrderByTime() (timeStamps []string, counts []int) {
	if tc.ordered {
		return tc.Timestamps, tc.Counts
	}
	sort.Strings(tc.Timestamps)
	for _, t := range tc.Timestamps {
		counts = append(counts, tc.Counter[t])
	}
	tc.ordered = true
	tc.Counts = counts
	return tc.Timestamps, counts
}

func (tc *TsCounter) OrderByCount() (timeStamps []string, counts []int) {
	panic("implement me")
}
