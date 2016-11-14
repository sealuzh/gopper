package data

import (
	"context"
	"fmt"
	"sort"
)

type TransFunc func(context.Context, <-chan *TestResult) <-chan *TestResult

func Transform(ctx context.Context, in Results, transformers ...TransFunc) Results {
	ret := NewResults()
	for _, r := range in.TestNames() {
		tests, ok := in.Get(r)
		if !ok {
			panic(fmt.Sprintf("No element in results with name '%s'", r))
		}

		ch := make(chan *TestResult)
		var c <-chan *TestResult = ch
		for _, transformer := range transformers {
			c = transformer(ctx, c)
		}
		ch <- tests
		// fmt.Printf("%d filtered results\n", len(tests.ExecutionResults))
		select {
		case results := <-c:
			// fmt.Printf("%d filtered results\n", len(results.ExecutionResults))
			if results != nil {
				for _, res := range results.ExecutionResults {
					ret.Add(res)
				}
			}
		case <-ctx.Done():
			break
		}
	}
	return ret
}

func MinVersions(v int) TransFunc {
	return func(ctx context.Context, in <-chan *TestResult) <-chan *TestResult {
		out := make(chan *TestResult)
		go func() {
			tests := <-in
			if tests == nil {
				out <- nil
				// fmt.Printf("MinVersion: in is nill\n")
				return
			}
			execResults := tests.ExecutionResults
			if len(execResults) >= v {
				out <- tests
			} else {
				out <- nil
				// fmt.Printf("MinVersion: too few results for '%s': %d\n", tests.Test, tests.Len())
			}
		}()
		return out
	}
}

func MinMeanRuntime(r float32) TransFunc {
	return func(ctx context.Context, in <-chan *TestResult) <-chan *TestResult {
		out := make(chan *TestResult)
		go func() {
			tests := <-in
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
			var avgRt float32 = 0.0
			for _, r := range execResults {
				avgRt += r.RawVal
			}
			avgRt = avgRt / float32(l)
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

func MinMedianRuntime(r float32) TransFunc {
	return func(ctx context.Context, in <-chan *TestResult) <-chan *TestResult {
		out := make(chan *TestResult)
		go func() {
			tests := <-in
			if tests == nil {
				// fmt.Printf("MinMedianRuntime: in is nill\n")
				out <- nil
				return
			}

			copy := tests.copy()
			sort.Sort(copy)
			l := copy.Len()

			if l == 0 {
				out <- nil
				// fmt.Printf("MinMedianRuntime: l is 0\n")
				return
			}

			var median float32
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
