package data

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
)

const minCapacity = 50

func newResult(record []string) *Result {
	if len(record) < 6 {
		return nil
	}
	rawVal, err := strconv.ParseFloat(record[5], 32)
	if err != nil {
		fmt.Printf("Could not parse RawVal (%v) of record (%v:%v)", record[5], record[2], record[4])
		return nil
	}
	return &Result{
		record[0],
		record[1],
		record[2],
		record[3],
		record[4],
		float32(rawVal),
	}
}

type Result struct {
	Project       string
	Version       string
	SHA           string
	Configuration string
	Test          string
	RawVal        float32
}

func (r Result) AsStringArray() []string {
	return []string{
		r.Project,
		r.Version,
		r.SHA,
		r.Configuration,
		r.Test,
		strconv.FormatFloat(float64(r.RawVal), 'f', -1, 32),
	}
}

// Results
type Results interface {
	Add(r *Result) error
	Get(test string) (testResults []*Result, ok bool)
	TestNames() []string
	Length() int
}

func NewResults() Results {
	return &resultsMap{
		m:     make(map[string][]*Result),
		names: make([]string, 0, minCapacity),
	}
}

func ResultsFromFile(path string) (heading []string, data Results, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	// assume csv file
	r := csv.NewReader(f)
	r.Comma = rune(';')
	r.Comment = rune('#')
	r.LazyQuotes = true

	res := NewResults()
	// ignore first line
	heading, err = r.Read()
	if err != nil {
		return nil, nil, err
	}
	for {
		rec, err := r.Read()
		if err != nil {
			if err != io.EOF {
				fmt.Printf("ERROR - could not read CSV line: %v\n", err)
			}
			break
		}
		result := newResult(rec)

		if result == nil {
			continue
		}
		res.Add(result)
	}

	if err != nil {
		return nil, nil, err
	}
	return heading, res, nil
}

type resultsMap struct {
	lock  sync.RWMutex
	m     map[string][]*Result
	names []string
}

func (rm *resultsMap) Add(r *Result) error {
	if r == nil {
		return fmt.Errorf("Result to add is nil")
	}

	rm.lock.Lock()
	defer rm.lock.Unlock()
	res, ok := rm.m[r.Test]
	if ok {
		rm.m[r.Test] = append(res, r)
		rm.names = append(rm.names, r.Test)
	} else {
		rm.m[r.Test] = append(make([]*Result, 0, minCapacity), r)
		rm.names = append(rm.names, r.Test)
	}

	return nil
}

func (rm *resultsMap) Get(test string) ([]*Result, bool) {
	rm.lock.RLock()
	e, ok := rm.m[test]
	rm.lock.RUnlock()
	if !ok {
		return nil, false
	}
	return e, true
}

func (rm *resultsMap) Length() int {
	rm.lock.RLock()
	defer rm.lock.RUnlock()
	return len(rm.m)
}

func (rm *resultsMap) TestNames() []string {
	rm.lock.RLock()
	defer rm.lock.RUnlock()
	ret := make([]string, len(rm.m))
	copy(ret, rm.names)
	return ret
}
