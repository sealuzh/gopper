package analyse

import (
	"context"
	"fmt"
	"io/ioutil"
	"reflect"

	"bitbucket.org/sealuzh/gopper/data"
)

func Bcp(script string, probability float64) (data.AnalysisFunc, error) {
	rm := newLocalRManager()
	b, err := ioutil.ReadFile(script)
	if err != nil {
		return nil, err
	}
	f := string(b)
	return func(ctx context.Context, tr data.TestResult) (data.ChangePoints, error) {
		if tr == nil {
			return nil, fmt.Errorf("Bcp function: parameter tr is nil")
		}

		res, err := rm.evaluate(tr, f)
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
				commit := tr.Commits()[i]
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
