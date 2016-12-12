package filter

import (
	"context"

	"bitbucket.org/sealuzh/gopper/data"
)

func MinMeanRuntime(r float64) data.TransFunc {
	return func(ctx context.Context, in <-chan *data.TestResult) <-chan *data.TestResult {
		out := make(chan *data.TestResult)
		go func() {
			defer close(out)
			tests, ok := <-in
			if !ok {
				return
			}
			if tests == nil {
				out <- nil
				// fmt.Printf("MinMeanRuntime: in is nill\n")
				return
			}

			l := tests.Len()

			if l == 0 {
				out <- nil
				// fmt.Printf("MinMeanRuntime: length is 0\n")
				return
			}

			execResults := tests.ExecutionResults
			var avgRt float64
			for _, r := range execResults {
				avgRt += r.RawVal
			}
			avgRt = avgRt / float64(l)
			if avgRt < r {
				out <- nil
				// fmt.Printf("MinMeanRuntime: below avg runtime '%s': %v\n", tests.Test, avgRt)
			} else {
				out <- tests
			}
		}()
		return out
	}
}
