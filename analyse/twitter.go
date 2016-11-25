package analyse

import (
	"context"
	"fmt"
	"io/ioutil"
	"reflect"

	"bitbucket.org/sealuzh/gopper/data"
)

const (
	rvarMinMean = "minMean"
)

func Twitter(script string, minMean int) (data.AnalysisFunc, error) {
	rm := newLocalRManager()
	s, err := ioutil.ReadFile(script)
	if err != nil {
		return nil, err
	}
	f := string(s)
	return func(ctx context.Context, tr *data.TestResult) (data.ChangePoints, error) {
		if tr == nil {
			return nil, fmt.Errorf("Twitter function: parameter tr is nil")
		}

		res, err := rm.evaluate(tr, f, rParam{name: rvarMinMean, value: []int32{int32(minMean)}})
		if err != nil {
			return nil, err
		}

		var resTyped []int32
		switch rt := res.(type) {
		case []int32:
			resTyped = rt
		case int32:
			resTyped = []int32{rt}
		default:
			return nil, fmt.Errorf("Twitter function: r script returned wrong result type '%v'", reflect.TypeOf(rt))
		}

		cps := data.NewChangePoints()
		cpCount := 0
		ler := len(tr.ExecutionResults)
		for _, cp := range resTyped {
			cp := int(cp)
			if cp > ler {
				return nil, fmt.Errorf("Twitter function: change point (%d) is out of range (%d)", cp, ler)
			}
			newCp, err := data.NewChangePoint(tr.ExecutionResults[cp-1])
			if err != nil {
				return nil, err
			}
			cps.Add(newCp)
			cpCount++
		}
		fmt.Printf("  %d change points in %s\n", cpCount, tr.Test)
		return cps, nil
	}, nil
}
