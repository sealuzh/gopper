package data

import (
	"context"
	"fmt"
)

type TransFunc func(context.Context, <-chan TestResult) <-chan TestResult

func Transform(ctx context.Context, in TestResults, transformers ...TransFunc) TestResults {
	ret := NewTestResults(in.Heading())
	for _, r := range in.TestNames() {
		tests, ok := in.Get(r)
		if !ok {
			panic(fmt.Sprintf("No element in results with name '%s'", r))
		}

		ch := make(chan TestResult)
		var c <-chan TestResult = ch
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
					ret.AddTest(results)
				}
			}
		case <-ctx.Done():
			break
		}
	}
	fmt.Printf("  %d/%d not filtered\n", ret.Len(), in.Len())
	return ret
}
