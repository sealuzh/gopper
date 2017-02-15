package data

import (
	"fmt"
	"strconv"
	"sync"
)

func newExecutionResult(record []string) *ExecutionResult {
	if len(record) < 6 {
		return nil
	}
	rawVal, err := strconv.ParseFloat(record[5], 64)
	if err != nil {
		fmt.Printf("Could not parse RawVal (%v) of record (%v:%v)", record[5], record[2], record[4])
		return nil
	}
	return &ExecutionResult{
		record[0],
		record[1],
		record[2],
		record[3],
		record[4],
		rawVal,
	}
}

type ExecutionResult struct {
	Project       string
	Version       string
	SHA           string
	Configuration string
	Test          string
	RawVal        float64
}

func (r ExecutionResult) AsStringArray() []string {
	return []string{
		r.Project,
		r.Version,
		r.SHA,
		r.Configuration,
		r.Test,
		strconv.FormatFloat(float64(r.RawVal), 'f', -1, 64),
	}
}

// ExecutionResults
const (
	defaultExecutionResultCount = 30
)

type ExecutionResults interface {
	Values() []float64
	All() []*ExecutionResult
	Add(er *ExecutionResult)
}

func newExecutionResults() ExecutionResults {
	return &executionResultsImpl{
		ers: make([]*ExecutionResult, 0, defaultExecutionResultCount),
	}
}

type executionResultsImpl struct {
	l   sync.RWMutex
	ers []*ExecutionResult
}

func (ers *executionResultsImpl) Values() []float64 {
	ers.l.RLock()
	defer ers.l.RUnlock()
	l := len(ers.ers)
	ret := make([]float64, l)
	for i, er := range ers.ers {
		ret[i] = er.RawVal
	}
	return ret
}

func (ers *executionResultsImpl) All() []*ExecutionResult {
	ers.l.RLock()
	defer ers.l.RUnlock()
	return ers.ers
}

func (ers *executionResultsImpl) Add(er *ExecutionResult) {
	ers.l.Lock()
	defer ers.l.Unlock()
	ers.ers = append(ers.ers, er)
}
