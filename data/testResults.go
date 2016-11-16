package data

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sync"
)

const minCapacity = 50
const sep = ';'
const comment = '#'

// Results
type TestResults interface {
	Add(r *ExecutionResult) error
	AddTest(t *TestResult) error
	Remove(test string) error
	Get(test string) (testResults *TestResult, ok bool)
	TestNames() []string
	Length() int
	Heading() []string
	HeadingString() string
}

func NewTestResults() TestResults {
	return NewTestResultsWithHeading([]string{})
}

func NewTestResultsWithHeading(heading []string) TestResults {
	return &testResultsMap{
		m:       make(map[string]*TestResult),
		names:   make([]string, 0, minCapacity),
		heading: heading,
	}
}

func TestResultsFromFile(path string) (data TestResults, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// assume csv file
	r := csv.NewReader(f)
	r.Comma = rune(sep)
	r.Comment = rune(comment)
	r.LazyQuotes = true

	// ignore first line
	heading, err := r.Read()
	if err != nil {
		return nil, err
	}
	res := NewTestResultsWithHeading(heading)
	for {
		rec, err := r.Read()
		if err != nil {
			if err != io.EOF {
				fmt.Printf("ERROR - could not read CSV line: %v\n", err)
			}
			break
		}
		result := newExecutionResult(rec)

		if result == nil {
			continue
		}
		res.Add(result)
	}

	if err != nil {
		return nil, err
	}
	return res, nil
}

type testResultsMap struct {
	lock    sync.RWMutex
	m       map[string]*TestResult
	names   []string
	heading []string
}

func (rm *testResultsMap) Heading() []string {
	return rm.heading
}

func (rm *testResultsMap) HeadingString() string {
	ret := bytes.Buffer{}
	h := rm.Heading()
	for i, e := range h {
		ret.WriteString(e)
		if i < (len(h) - 1) {
			ret.WriteRune(sep)
		}
	}
	return ret.String()
}

func (rm *testResultsMap) Add(r *ExecutionResult) error {
	if r == nil {
		return fmt.Errorf("Result to add is nil")
	}

	rm.lock.Lock()
	defer rm.lock.Unlock()
	res, ok := rm.m[r.Test]
	if ok {
		res.ExecutionResults = append(res.ExecutionResults, r)
	} else {
		rm.m[r.Test] = &TestResult{
			ExecutionResults: append(make([]*ExecutionResult, 0, minCapacity), r),
			Project:          r.Project,
			Test:             r.Test,
		}
		rm.names = append(rm.names, r.Test)
	}

	return nil
}

func (rm *testResultsMap) AddTest(t *TestResult) error {
	if t == nil {
		return fmt.Errorf("Test to add is nil")
	}

	rm.lock.Lock()
	defer rm.lock.Unlock()

	res, ok := rm.m[t.Test]
	if ok {
		res.ExecutionResults = append(res.ExecutionResults, t.ExecutionResults...)
	} else {
		rm.m[t.Test] = t
		rm.names = append(rm.names, t.Test)
	}

	return nil
}

func (rm *testResultsMap) Remove(test string) error {
	rm.lock.Lock()
	defer rm.lock.Unlock()

	fmt.Printf("Size of results: %d; names: %d\n", len(rm.m), len(rm.names))
	_, ok := rm.m[test]
	if !ok {
		return fmt.Errorf("No element with name '%s' to remove", test)
	}
	delete(rm.m, test)

	index := -1
	for _, n := range rm.names {
		index += 1
		if n == test {
			fmt.Printf("Remove element name: %s\n", n)
			break
		}
	}

	if index != -1 {
		if index < len(rm.names) {
			rm.names = append(rm.names[:index], rm.names[index+1:]...)
		} else {
			rm.names = rm.names[:index]
		}
	} else {
		fmt.Printf("Could not find name: %s\n", test)
	}
	fmt.Printf("Size of results: %d; names: %d\n", len(rm.m), len(rm.names))

	return nil
}

func (rm *testResultsMap) Get(test string) (*TestResult, bool) {
	rm.lock.RLock()
	e, ok := rm.m[test]
	rm.lock.RUnlock()
	if !ok {
		return nil, false
	}
	return e, true
}

func (rm *testResultsMap) Length() int {
	rm.lock.RLock()
	defer rm.lock.RUnlock()
	return len(rm.m)
}

func (rm *testResultsMap) TestNames() []string {
	rm.lock.RLock()
	defer rm.lock.RUnlock()
	ret := make([]string, len(rm.m))
	copy(ret, rm.names)
	return ret
}
