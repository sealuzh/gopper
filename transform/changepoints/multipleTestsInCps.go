package changepoints

import (
	"context"
	"fmt"

	"bitbucket.org/sealuzh/gopper/data"
)

func RemoveDuplicateTestNames(ctx context.Context, cps []data.ChangePoints) ([]data.ChangePoints, error) {
	tns := make(map[string]struct{})
	ret := make([]data.ChangePoints, 0, len(cps))
	for i, cp := range cps {
		cps := cp.All()
		startLen := cp.Len()
		newCp := data.NewChangePoints()
		for _, c := range cps {
			for _, tn := range c.TestNames() {
				er, ok := c.Get(tn)
				if !ok {
					panic(fmt.Sprintf("Inconsistent change point. commit=%s; test=%s", c.Commit(), tn))
				}

				key := tn + er.SHA
				if _, ok := tns[key]; !ok {
					newC, err := data.NewChangePoint(er)
					if err != nil {
						return nil, err
					}
					err = newCp.Add(newC)
					if err != nil {
						return nil, err
					}
					tns[key] = struct{}{}
				}
			}
		}
		endLen := newCp.Len()
		if endLen > 0 {
			ret = append(ret, newCp)
		}
		fmt.Printf("  Set %d: %d/%d\n", i, endLen, startLen)
	}
	return ret, nil
}
