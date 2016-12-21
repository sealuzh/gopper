package data

import (
	"context"
	"fmt"
)

func Merge(ctx context.Context, ins []TestResults) TestResults {
	var r TestResults
	for i, in := range ins {
		if i == 0 {
			r = NewTestResults(in.Heading())
		}
		for _, n := range in.TestNames() {
			t, ok := in.Get(n)
			if !ok {
				panic(fmt.Sprintf("TestNames and Get inconsistent for name '%s'", n))
			}
			r.AddTest(t)
		}
	}
	return r
}
