package analyse

import "bitbucket.org/sealuzh/gopper/data"

func vectorise(r *data.TestResult) []float32 {
	l := len(r.ExecutionResults)
	ret := make([]float32, l)
	for i, r := range r.ExecutionResults {
		ret[i] = r.RawVal
	}
	return ret
}
