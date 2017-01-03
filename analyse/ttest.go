package analyse

import (
	"context"
	"fmt"
	"reflect"

	"bitbucket.org/sealuzh/gopper/data"
	"github.com/senseyeio/roger"
)

const (
	var1Name            = "v1"
	var2Name            = "v2"
	varPairedName       = "paired"
	expectedResultCount = 3
	falseString         = "FALSE"
	trueString          = "TRUE"
	ttestScript         = "require(\"stats\")\nres <- t.test(%s, %s, paired = %s)\nc(res$statistic, res$parameter, res$p.value)"
)

func Ttest(sig float64, paired bool) (data.AnalysisFunc, error) {
	rm := newLocalRManager()
	return func(ctx context.Context, tr data.TestResult) (data.ChangePoints, error) {
		if tr == nil {
			return nil, fmt.Errorf("Parameter tr is nil")
		}

		table := vectoriseAll(tr)
		lTable := len(table)

		cps := data.NewChangePoints()

		// check if there are at least 2 versions to compare
		if lTable < 2 {
			return cps, nil
		}

		commits := tr.Commits()
		cpCount := 0

		c := rm.client()

		for j := 1; j < lTable; j++ {
			i := j - 1
			resI := table[i]
			resJ := table[j]

			res, err := changes(c, true, resI, resJ)
			if err != nil {
				return nil, err
			}

			if res.pValue < 1-sig {
				cp, err := data.NewChangePoint(commits[i], tr)
				if err != nil {
					return nil, err
				}
				err = cps.Add(cp)
				if err != nil {
					return nil, err
				}
				cpCount++
			}
		}
		fmt.Printf("  %d change points in %s\n", cpCount, tr.Test())
		return cps, nil
	}, nil
}

type ttestResult struct {
	tStatistics      float64
	degreesOfFreedom float64
	pValue           float64
}

func changes(c roger.RClient, paired bool, var1, var2 []float64) (*ttestResult, error) {
	s, err := c.GetSession()
	defer s.Close()

	if err != nil {
		return nil, err
	}

	pairedToString := falseString
	if paired {
		pairedToString = trueString
	}

	res, err := s.Eval(fmt.Sprintf(ttestScript, f64SliceToString(var1), f64SliceToString(var2), pairedToString))
	if err != nil {
		return nil, err
	}

	switch r := res.(type) {
	case []float64:
		l := len(r)
		if l != expectedResultCount {
			return nil, fmt.Errorf("Ttest function: r script returned %d results, but expected %d", l, expectedResultCount)
		}
		return &ttestResult{
			tStatistics:      r[0],
			degreesOfFreedom: r[1],
			pValue:           r[2],
		}, nil
	default:
		return nil, fmt.Errorf("Ttest function: r script returned wrong result type '%v'", reflect.TypeOf(r))
	}
}
