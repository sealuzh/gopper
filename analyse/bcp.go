package analyse

import (
	"context"
	"fmt"
	"io/ioutil"

	"bitbucket.org/sealuzh/gopper/data"
)

func Bcp(script string) (data.AnalysisFunc, error) {
	rm := newRManager()
	b, err := ioutil.ReadFile(script)
	if err != nil {
		return nil, err
	}
	f := string(b)
	fmt.Println(f)
	return func(ctx context.Context, tr *data.TestResult) ([]string, error) {
		if tr == nil {
			return nil, fmt.Errorf("Bcp function: parameter tr is nil")
		}

		rc := rm.client()
		s, err := rc.GetSession()
		if err != nil {
			return nil, err
		}

		s.Assign("td", vectorise(tr))
		res, err := s.Eval(f)
		if err != nil {
			return nil, err
		}

		//TODO
		fmt.Printf("-----------\n%v\n------------", res)

		panic("analyse.Bcp NOT IMPLEMENTED")
	}, nil
}
