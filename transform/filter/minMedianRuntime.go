package filter

import (
	"context"
	"sort"

	"bitbucket.org/sealuzh/gopper/data"
)

func MinMedianRuntime(r float64) data.TransFunc {
	return func(ctx context.Context, in <-chan *data.TestResult) <-chan *data.TestResult {
		out := make(chan *data.TestResult)
		go func() {
			defer close(out)
			tests, ok := <-in
			if !ok {
				return
			}

			if tests == nil {
				// fmt.Printf("MinMedianRuntime: in is nill\n")
				out <- nil
				return
			}

			copy := tests.Copy()
			sort.Sort(copy)
			l := copy.Len()

			if l == 0 {
				out <- nil
				// fmt.Printf("MinMedianRuntime: l is 0\n")
				return
			}

			var median float64
			if l == 1 {
				median = copy.ExecutionResults[0].RawVal
			} else if l%2 != 0 {
				median = (copy.ExecutionResults[l/2].RawVal + copy.ExecutionResults[l/2].RawVal + 1) / 2
			} else {
				median = copy.ExecutionResults[l/2].RawVal
			}

			if median < r {
				out <- nil
				// fmt.Printf("MinMedianRuntime: below median '%s': %v\n", tests.Test, median)
			} else {
				out <- tests
			}
		}()
		return out
	}
}
