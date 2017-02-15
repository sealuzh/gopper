package data

import (
	"fmt"
	"sync"
)

const (
	defaultExecutionResultLength = 30
)

type TestResult interface {
	Project() string
	Test() string
	Commits() []string
	ExecutionResult(commit string) (ExecutionResults, bool)
	AddExecutionResult(er *ExecutionResult) error
	ChangePoints() ChangePoints
	AddChangePoint(cp ChangePoint) error
	Copy() TestResult
}

func NewTestResult(project, test string) TestResult {
	return &testResultImpl{
		project:          project,
		test:             test,
		executionResults: make(map[string]ExecutionResults),
		commits:          make([]string, 0, defaultExecutionResultLength),
		changePoints:     NewChangePoints(),
	}
}

type testResultImpl struct {
	l                sync.RWMutex
	project          string
	test             string
	commits          []string
	executionResults map[string]ExecutionResults
	changePoints     ChangePoints
}

func (t *testResultImpl) Project() string {
	// no locking as t.project is effectively immutable
	return t.project
}

func (t *testResultImpl) Test() string {
	// no locking as t.test is effectively immutable
	return t.test
}

func (t *testResultImpl) Commits() []string {
	t.l.RLock()
	defer t.l.RUnlock()
	l := len(t.commits)
	c := make([]string, len(t.commits))
	copiedElements := copy(c, t.commits)
	if l != copiedElements {
		panic(fmt.Sprintf("testResultImpl: only copied %d of %d elements", copiedElements, l))
	}
	return c
}

func (t *testResultImpl) ExecutionResult(commit string) (ExecutionResults, bool) {
	t.l.RLock()
	defer t.l.RUnlock()
	er, ok := t.executionResults[commit]
	return er, ok
}

func (t *testResultImpl) AddExecutionResult(er *ExecutionResult) error {
	if er == nil {
		return fmt.Errorf("Parameter er is nil")
	}
	t.l.Lock()
	defer t.l.Unlock()

	var contained bool
	for _, c := range t.commits {
		if c == er.SHA {
			contained = true
			break
		}
	}

	if contained {
		ers, ok := t.executionResults[er.SHA]
		if !ok {
			panic(fmt.Sprintf("testResultImpl::AddExecutionResult - Incorrect state of commits and executionReuslts: %v", er.SHA))
		}
		t.executionResults[er.SHA] = append(ers, er)
	} else {
		t.commits = append(t.commits, er.SHA)
		t.executionResults[er.SHA] = append(make(ExecutionResults, 0, defaultExecutionResultLength), er)
	}
	return nil
}

func (t *testResultImpl) ChangePoints() ChangePoints {
	t.l.RLock()
	defer t.l.RUnlock()
	return t.changePoints.Copy()
}

func (t *testResultImpl) AddChangePoint(cp ChangePoint) error {
	t.l.Lock()
	defer t.l.Unlock()
	return t.changePoints.Add(cp)
}

func (t *testResultImpl) Copy() TestResult {
	t.l.RLock()
	defer t.l.RUnlock()

	lc := len(t.commits)
	commits := make([]string, lc)
	copiedCommits := copy(commits, t.commits)
	if lc != copiedCommits {
		panic(fmt.Sprintf("testResultImpl: only copied %d of %d elements", copiedCommits, lc))
	}

	exRes := make(map[string]ExecutionResults)
	for k, v := range t.executionResults {
		exRes[k] = v
	}

	return &testResultImpl{
		project:          t.project,
		test:             t.test,
		commits:          commits,
		executionResults: exRes,
		changePoints:     t.changePoints.Copy(),
	}
}
