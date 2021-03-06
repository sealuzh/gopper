package filter

import (
	"context"
	"fmt"

	"github.com/montanaflynn/stats"
	"github.com/sealuzh/gopper/data"
)

func MinMedianRuntime(r float64) data.TransFunc {
	return func(ctx context.Context, in <-chan data.TestResult) <-chan data.TestResult {
		out := make(chan data.TestResult)
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

			commits := tests.Commits()
			medians := make([]float64, len(commits))
			for i, c := range commits {
				ers, ok := tests.ExecutionResults(c)
				if !ok {
					panic(fmt.Sprintf("Inconsistent test result: %s", c))
				}
				median, err := stats.Median(stats.Float64Data(ers.Values()))
				if err != nil {
					panic(err)
				}
				medians[i] = median
			}

			median, err := stats.Median(stats.Float64Data(medians))
			if err != nil {
				panic(err)
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
