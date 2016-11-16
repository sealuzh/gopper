package data

import (
	"context"
	"fmt"
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
		case results, ok := <-c:
			// fmt.Printf("%d filtered results\n", len(results.ExecutionResults))
			if ok {
				if results != nil {
					for _, res := range results.ExecutionResults {
						ret.Add(res)
					}
				}
			}
		case <-ctx.Done():
			break
		}
	}
	return ret
}
