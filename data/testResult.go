package data

const (
	defaultExecutionResultLength = 30
)

type TestResult struct {
	ExecutionResults []*ExecutionResult
	Project          string
	Test             string
	ChangePoints     ChangePoints
}

func (t TestResult) Copy() *TestResult {
	exRes := make([]*ExecutionResult, len(t.ExecutionResults))
	copy(exRes, t.ExecutionResults)

	var cps ChangePoints
	if t.ChangePoints == nil {
		cps = NewChangePoints()
	} else {
		cps = t.ChangePoints.Copy()
	}

	return &TestResult{
		Project:          t.Project,
		Test:             t.Test,
		ExecutionResults: exRes,
		ChangePoints:     cps,
	}
}

func (t TestResult) Len() int {
	return len(t.ExecutionResults)
}

func (t TestResult) Less(i, j int) bool {
	return t.ExecutionResults[i].RawVal <= t.ExecutionResults[j].RawVal
}

func (t *TestResult) Swap(i, j int) {
	buffer := t.ExecutionResults[i]
	t.ExecutionResults[i] = t.ExecutionResults[j]
	t.ExecutionResults[j] = buffer
}
