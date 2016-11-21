package analyse

import "bitbucket.org/sealuzh/gopper/data"

func vectorise(r *data.TestResult) []float64 {
	l := len(r.ExecutionResults)
	ret := make([]float64, l)
	for i, r := range r.ExecutionResults {
		ret[i] = float64(r.RawVal)
	}
	return ret
}
