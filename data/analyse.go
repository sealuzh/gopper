package data

import (
	"context"
	"fmt"
	"sync"
)

const (
	maxWorkers = 10
)

type AnalysisFunc func(context.Context, TestResult) (ChangePoints, error)

func Analyse(ctx context.Context, in TestResults, f AnalysisFunc) TestResults {
	res := make(chan TestResult)
	out := make(chan TestResult)
	var wg sync.WaitGroup

	tns := in.TestNames()
	ltns := len(tns)
	wc := workerCount(ltns)
	for i := 0; i < wc; i++ {
		go runAnalysis(ctx, f, out, res)
	}
	wg.Add(ltns)
	ret := NewTestResults()
	for _, n := range tns {
		n := n
		r, ok := in.Get(n)
		if !ok {
			panic(fmt.Sprintf("TestNames and Get inconsistent for name '%s'\n", n))
		}
		go func() {
			select {
			case out <- r:
			case <-ctx.Done():
			}
		}()
	}

	// goroutine that closes the c channel
	go func() {
		wg.Wait()
		close(res)
	}()

	for tr := range res {
		ret.AddTest(tr)
		wg.Done()
	}
	close(out)
	return ret
}

func runAnalysis(ctx context.Context, f AnalysisFunc, in <-chan TestResult, res chan<- TestResult) {
	i := in
	var c chan<- TestResult
	var tr TestResult

Loop:
	for {
		select {
		case r, ok := <-i:
			// receive results on in channel
			if !ok {
				break Loop
			}
			cps, err := f(ctx, r)
			if err != nil {
				if err != context.Canceled {
					fmt.Printf("ERROR - analysis function returned with an error for '%s': %v\n", r.Test, err)
					tr = nil
					c = res
					i = nil
					break
				}
			}
			for _, cp := range cps.All() {
				r.AddChangePoint(cp)
			}

			c = res
			i = nil
			tr = r
		case c <- tr:
			// send result on result channel
			c = nil
			i = in
			continue
		case <-ctx.Done():
			// wait if context gets canceled
			break Loop
		}
	}
}

func workerCount(r int) int {
	if r < maxWorkers {
		return r
	}
	return maxWorkers
}
