package filter

import (
	"context"

	"bitbucket.org/sealuzh/gopper/data"
)

func MinVersions(v int) data.TransFunc {
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
				// fmt.Printf("MinVersion: in is nill\n")
				return
			}
			commits := tests.Commits()
			if len(commits) >= v {
				out <- tests
			} else {
				out <- nil
				// fmt.Printf("MinVersion: too few results for '%s': %d\n", tests.Test, tests.Len())
			}
		}()
		return out
	}
}
