package analyse

import (
	"context"
	"fmt"
	"io/ioutil"
	"reflect"

	"bitbucket.org/sealuzh/gopper/data"
)

const (
	rvarTestData = "td"
)

func Bcp(script string, probability float64) (data.AnalysisFunc, error) {
	rm := newLocalRManager()
	b, err := ioutil.ReadFile(script)
	if err != nil {
		return nil, err
	}
	f := string(b)
	return func(ctx context.Context, tr *data.TestResult) ([]string, error) {
		if tr == nil {
			return nil, fmt.Errorf("Bcp function: parameter tr is nil")
		}

		rc := rm.client()
		s, err := rc.GetSession()
		if err != nil {
			return nil, err
		}
		defer s.Close()

		data := vectorise(tr)
		err = s.Assign(rvarTestData, data)
		if err != nil {
			return nil, fmt.Errorf("Bcp function: could not assign test data")
		}
		res, err := s.Eval(f)
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

		lcps := len(cps)
		ler := len(tr.ExecutionResults)
		if lcps != ler {
			return nil, fmt.Errorf("Bcp functions: returned change points (%d) not equal to execution results (%d)", lcps, ler)
		}

		ret := make([]string, 0, lcps)
		for i, cp := range cps {
			if cp >= probability {
				ret = append(ret, tr.ExecutionResults[i].SHA)
			}
		}
		fmt.Printf("  %d change points in %s\n", len(ret), tr.Test)
		return ret, nil
	}, nil
}
