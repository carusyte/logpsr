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
	"io"
	"time"
)

type FileType string

const (
	GATEWAY  FileType = "gateway"
	UC_AS_T2 FileType = "ucas_t2"
)

func main() {
	start, end := "14:00:00", "18:18:00"
	fmap := map[FileType]string{
		//GATEWAY:  "/Users/jx/Downloads/ctcs-gateway.log.2017-10-20",
		UC_AS_T2: "/Users/jx/Downloads/ucas.log.1020/t2traces.log.2017-10-20",
	}
	tcs := make([]*TsCounter, 0, 16)
	for n, f := range fmap {
		log.Printf("parsing file %s", f)
		tc, nlns, e := ParseFile(n, f, start, end)
		log.Printf("%+v lines parsed: %d", f, nlns)
		if e != nil {
			log.Panicln(e)
		}
		tc.Even(start, end)
		tc.OrderByTime()
		tcs = append(tcs, tc)
	}

	graph("points_1020.png", tcs...)
}

func graph(dstFile string, tcs ... *TsCounter) {
	log.Printf("generating plot graph...")

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

func ParseFile(ftype FileType, filePath, start, end string) (tc *TsCounter, nlines int, e error) {
	inclSt, inclEd := true, true
	if strings.HasPrefix(start, "(") {
		inclSt = false
		start = start[1:]
	}
	if strings.HasSuffix(end, ")") {
		inclEd = false
		end = end[:len(end)-1]
	}

	tc = new(TsCounter)
	tc.Name = string(ftype)
	tag := ""
	switch ftype {
	case GATEWAY:
		tag = `"timestamp":".{12}(.{8}).{4}",`
	case UC_AS_T2:
		tag = `.{11}(.{8}),.*current funtionId:306122, parsing request parameters`
	}
	rex := regexp.MustCompile(tag)

	inputFile, err := os.Open(filePath)
	if err != nil {
		return nil, 0, err
	}
	defer inputFile.Close()

	rd := bufio.NewReader(inputFile)
	s, e := Readln(rd)
	for e == nil {
		nlines++
		r := rex.FindStringSubmatch(s)
		if len(r) > 0 {
			timeVal := r[len(r)-1]
			if start != "" {
				if inclSt {
					if start <= timeVal {
						tc.Add(timeVal)
					}
				} else {
					if start < timeVal {
						tc.Add(timeVal)
					}
				}
			} else {
				tc.Add(timeVal)
			}
			if end != "" {
				if inclEd {
					if timeVal <= end {
						tc.Add(timeVal)
					} else {
						break
					}
				} else {
					if timeVal < end {
						tc.Add(timeVal)
					} else {
						break
					}
				}
			} else {
				tc.Add(timeVal)
			}
		}
		s, e = Readln(rd)
	}
	if e != nil && e != io.EOF {
		return nil, 0, e
	}

	return
}

func ParseLines(filePath string) ([]string, error) {
	inputFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer inputFile.Close()

	rd := bufio.NewReader(inputFile)
	var results []string
	s, e := Readln(rd)
	for e == nil {
		results = append(results, s)
		s, e = Readln(rd)
	}
	if e != nil && e != io.EOF {
		return nil, e
	}
	return results, nil
}

// Readln returns a single line (without the ending \n)
// from the input buffered reader.
// An error is returned iff there is an error with the
// buffered reader.
func Readln(r *bufio.Reader) (string, error) {
	var (
		isPrefix       = true
		err      error = nil
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return string(ln), err
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

func (tc *TsCounter) Even(start, end string) {
	inclSt, inclEd := true, true
	tf := "00:00:00"
	if strings.HasPrefix(start, "(") {
		inclSt = false
		start = start[1:]
	}
	if strings.HasSuffix(end, ")") {
		inclEd = false
		end = end[:len(end)-1]
	}

	st, e := time.Parse(tf, start)
	if e != nil {
		log.Panicln(e)
	}
	et, e := time.Parse(tf, end)
	if e != nil {
		log.Panicln(e)
	}
	if !inclSt {
		st.Add(time.Second)
	}
	if !inclEd {
		et.Add(-time.Second)
	}
	for st.Before(et) {
		t := st.Format(tf)
		if _, ok := tc.Counter[t]; !ok {
			tc.Counter[t] = 0
			tc.Timestamps = append(tc.Timestamps, t)
		}
		st.Add(time.Second)
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
