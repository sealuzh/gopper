package transform

import (
	"context"
	"fmt"

	"bitbucket.org/sealuzh/gopper/data"
)

func TestResultsToChangePoints(ctx context.Context, trs data.TestResults) (data.ChangePoints, error) {
	if trs == nil {
		return nil, fmt.Errorf("Parameter trs is nil")
	}

	cps := data.NewChangePoints()
	for tr := range trs.All() {
		for _, er := range tr.ExecutionResults {
			cp, ok := cps.Get(er.SHA)
			if ok {
				// change point for commit already exists -> add this execution result to the change point
				err := cp.Add(er)
				if err != nil {
					return nil, err
				}
			} else {
				// change point for commit does not exist yet -> create the change point
				cp, err := data.NewChangePoint(er)
				if err != nil {
					return nil, err
				}
				err = cps.Add(cp)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return cps, nil
}
