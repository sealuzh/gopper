package testresults

import (
	"context"
	"fmt"

	"bitbucket.org/sealuzh/gopper/data"
)

func ToChangePoints(ctx context.Context, trs data.TestResults) (data.ChangePoints, error) {
	if trs == nil {
		return nil, fmt.Errorf("Parameter trs is nil")
	}

	cps := data.NewChangePoints()
	for tr := range trs.All() {
		for _, c := range tr.ChangePoints().All() {
			cps.Add(c)
		}
	}
	fmt.Printf("  %d change points in %d tests\n", cps.Len(), trs.Len())
	return cps, nil
}
