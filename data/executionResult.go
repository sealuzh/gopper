package data

import (
	"fmt"
	"strconv"
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

func ExecutionResultsToValues(ers ExecutionResults) []float64 {
	ret := make([]float64, len(ers))
	for i, er := range ers {
		ret[i] = er.RawVal
	}
	return ret
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

type ExecutionResults []*ExecutionResult
