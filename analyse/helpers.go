package analyse

import "bitbucket.org/sealuzh/gopper/data"

func vectoriseFirstElement(r data.TestResult) []float64 {
	commits := r.Commits()
	l := len(commits)
	ret := make([]float64, l)
	for i, c := range commits {
		ers, ok := r.ExecutionResult(c)
		if !ok {
			panic("Incorrect TestResult state")
		}
		ret[i] = float64(ers[0].RawVal)
	}
	return ret
}
