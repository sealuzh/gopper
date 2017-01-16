package data

import (
	"fmt"

	"github.com/montanaflynn/stats"
)

const (
	jsonImprovement = "improvement: %d - %d %%"
	jsonRegression  = "regression: %d - %d %%"
)

type ChangePointType interface {
	fmt.Stringer
	IsRegression() bool
	IsImprovement() bool
	Category() ChangeCategory
}

func AllChangePointTypes() []ChangePointType {
	ret := make([]ChangePointType, 0, (2 * changeCategoryCount))
	cats := allChangeCategories()
	for _, c := range cats {
		ret = append(ret, cpt{
			regression: true,
			category:   c,
		})
	}
	for _, c := range cats {
		ret = append(ret, cpt{
			regression: false,
			category:   c,
		})
	}
	return ret
}

func NewDefaultChangePointType(regression bool, category ChangeCategory) ChangePointType {
	return cpt{
		regression: regression,
		category:   category,
	}
}

func ChangePointTypeFromResult(commit string, testResult TestResult) (ChangePointType, error) {
	commits := testResult.Commits()
	l := len(commits)
	var commit2 string
	for i, c := range commits {
		if c == commit {
			if i == (l - 1) {
				// last commit in test result
				fmt.Printf("ERROR - ChangePoint as last commit (%s) of test result (%s)\n", c, testResult.Test())
				// should never happen
				return cpt{}, nil
			}
			commit2 = commits[i+1]
			break
		}
	}

	// compare means
	trs1, ok := testResult.ExecutionResult(commit)
	if !ok {
		return nil, fmt.Errorf("NewChangePointType - No execution results for commit: %s", commit)
	}
	c1Data := make([]float64, len(trs1))
	for i, er := range trs1 {
		c1Data[i] = er.RawVal
	}
	c1Mean, err := stats.Mean(stats.Float64Data(c1Data))
	if err != nil {
		return nil, fmt.Errorf("NewChangePointType - error calculating mean: %v", err)
	}

	trs2, ok := testResult.ExecutionResult(commit2)
	if !ok {
		return nil, fmt.Errorf("NewChangePointType - No execution results for commit: %s", commit2)
	}
	c2Data := make([]float64, len(trs2))
	for i, er := range trs2 {
		c2Data[i] = er.RawVal
	}
	c2Mean, err := stats.Mean(stats.Float64Data(c2Data))
	if err != nil {
		return nil, fmt.Errorf("NewChangePointType - error calculating mean: %v", err)
	}

	cc := calcChangeCategory(c1Mean, c2Mean)
	return cpt{
		regression: c1Mean < c2Mean,
		category:   cc,
	}, nil
}

type cpt struct {
	regression bool
	category   ChangeCategory
}

func (t cpt) IsRegression() bool {
	return t.regression
}

func (t cpt) IsImprovement() bool {
	return !t.regression
}

func (t cpt) Category() ChangeCategory {
	return t.category
}

func (t cpt) String() string {
	from := t.category * 10
	to := from + 9
	if t.IsImprovement() {
		return fmt.Sprintf(jsonImprovement, from, to)
	} else if t.IsRegression() {
		return fmt.Sprintf(jsonRegression, from, to)
	}
	return fmt.Sprintf("equal")
}

func (t cpt) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

type ChangePointTypeError error
