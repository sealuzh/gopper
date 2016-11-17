package data

import (
	"context"
	"fmt"
	"sync"
)

type AnalysisFunc func(context.Context, *TestResult) ([]string, error)

func Analyse(ctx context.Context, in TestResults, f AnalysisFunc) TestResults {
	c := make(chan *TestResult)
	var wg sync.WaitGroup

	for _, n := range in.TestNames() {
		n := n
		r, ok := in.Get(n)
		if !ok {
			panic(fmt.Sprintf("TestNames and Get inconsistent for name '%s'\n", n))
		}
		wg.Add(1)
		go func() {
			// send done after goroutine returns
			defer func() { wg.Done() }()
			cps, err := f(ctx, r)
			if err != nil {
				if err != context.Canceled {
					fmt.Printf("ERROR - analysis function returned with an error for '%s': %v\n", n, err)
					return
				}
			}
			r.ChangePoints = cps
			select {
			case c <- r:
			case <-ctx.Done():
			}
		}()
	}

	// goroutine that closes the c channel
	go func() {
		wg.Wait()
		close(c)
	}()

	ret := NewTestResults()
	for tr := range c {
		ret.AddTest(tr)
	}
	return ret
}
