package data

type TestResult struct {
	ExecutionResults []*ExecutionResult
	Project          string
	Test             string
}

func (t TestResult) copy() *TestResult {
	exRes := make([]*ExecutionResult, len(t.ExecutionResults))
	copy(exRes, t.ExecutionResults)
	return &TestResult{
		Project:          t.Project,
		Test:             t.Test,
		ExecutionResults: exRes,
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
