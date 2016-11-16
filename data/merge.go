package data

import (
	"context"
	"fmt"
)

func Merge(ctx context.Context, ins []TestResults) TestResults {
	r := NewTestResults()
	for _, i := range ins {
		for _, n := range i.TestNames() {
			t, ok := i.Get(n)
			if !ok {
				panic(fmt.Sprintf("TestNames and Get inconsistent for name '%s'", n))
			}
			r.AddTest(t)
		}
	}
	return r
}
