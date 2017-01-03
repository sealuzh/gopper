package analyse

import (
	"context"
	"fmt"
	"reflect"

	"github.com/sealuzh/gopper/data"
)

const bcpScript = "library(\"bcp\")\ncp <- bcp(td)\ncp$posterior.prob"

func Bcp(probability float64) (data.AnalysisFunc, error) {
	rm := newLocalRManager()
	return func(ctx context.Context, tr data.TestResult) (data.ChangePoints, error) {
		if tr == nil {
			return nil, fmt.Errorf("Bcp function: parameter tr is nil")
		}

		res, err := rm.evaluate(tr, bcpScript)
		if err != nil {
			return nil, err
		}

		var cps []float64
		switch r := res.(type) {
		case []float64:
			cps = r
		default:
			return nil, fmt.Errorf("Bsp function: r script returned wrong result type '%v'", reflect.TypeOf(r))
		}

		commits := tr.Commits()
		lcps := len(cps)
		ler := len(commits)
		if lcps != ler {
			return nil, fmt.Errorf("Bcp functions: returned change points (%d) not equal to execution results (%d)", lcps, ler)
		}

		ret := data.NewChangePoints()
		cpCount := 0
		for i, cp := range cps {
			if cp >= probability {
				commit := commits[i]
				ncp, err := data.NewChangePoint(commit, tr)
				if err != nil {
					return nil, err
				}
				err = ret.Add(ncp)
				if err != nil {
					return nil, err
				}
				cpCount++
			}
		}
		fmt.Printf("  %d change points in %s\n", cpCount, tr.Test())
		return ret, nil
	}, nil
}
