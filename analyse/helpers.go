package analyse

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/sealuzh/gopper/data"
)

func vectoriseFirstElement(r data.TestResult) []float64 {
	commits := r.Commits()
	l := len(commits)
	ret := make([]float64, l)
	for i, c := range commits {
		ers, ok := r.ExecutionResults(c)
		if !ok {
			incorrectTestResultState(c, r)
		}
		ret[i] = float64(ers.All()[0].RawVal)
	}
	return ret
}

func vectoriseAll(r data.TestResult) [][]float64 {
	commits := r.Commits()
	lc := len(commits)
	ret := make([][]float64, lc)
	for i, c := range commits {
		ers, ok := r.ExecutionResults(c)
		if !ok {
			incorrectTestResultState(c, r)
		}

		ersData := ers.All()
		l := len(ersData)
		fErs := make([]float64, l)
		for i, er := range ersData {
			fErs[i] = er.RawVal
		}
		ret[i] = fErs
	}
	return ret
}

func incorrectTestResultState(commit string, tr data.TestResult) {
	panic(fmt.Sprintf("Incorrect test result state: %s @ %s", tr.Test(), commit))
}

func f64SliceToString(s []float64) string {
	var buf bytes.Buffer
	buf.WriteString("c(")
	l := len(s)
	for i, v := range s {
		sv := strconv.FormatFloat(v, 'f', -1, 64)
		buf.WriteString(sv)
		if i < l-1 {
			buf.WriteString(",")
		}
	}
	buf.WriteString(")")
	return buf.String()
}
