package data

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sync"
)

const minCapacity = 50

// Results
type Results interface {
	Add(r *ExecutionResult) error
	Remove(test string) error
	Get(test string) (testResults *TestResult, ok bool)
	TestNames() []string
	Length() int
}

func NewResults() Results {
	return &resultsMap{
		m:     make(map[string]*TestResult),
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
	m     map[string]*TestResult
	names []string
}

func (rm *resultsMap) Add(r *ExecutionResult) error {
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

func (rm *resultsMap) Remove(test string) error {
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

func (rm *resultsMap) Get(test string) (*TestResult, bool) {
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
