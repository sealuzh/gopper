package filter

import (
	"context"
	"fmt"

	"github.com/montanaflynn/stats"
	"github.com/sealuzh/gopper/data"
)

func MinMeanRuntime(r float64) data.TransFunc {
	return func(ctx context.Context, in <-chan data.TestResult) <-chan data.TestResult {
		out := make(chan data.TestResult)
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

			var avgRt float64
			counter := 0
			for _, c := range tests.Commits() {
				ers, ok := tests.ExecutionResults(c)
				if !ok {
					panic(fmt.Sprintf("Inconsistent test result: %s", c))
				}
				m, err := stats.Mean(stats.Float64Data(ers.Values()))
				if err != nil {
					panic(err)
				}
				avgRt += m
				counter++
			}
			avgRt = avgRt / float64(counter)

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
