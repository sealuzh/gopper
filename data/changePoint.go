package data

import (
	"fmt"
	"sync"
)

// ChangePoint
type ChangePoint interface {
	TestNames() []string
	Commit() string
	Type() ChangePointType
	Add(commit string, test TestResult) error
	Get(testName string) (TestResult, bool)
	Copy() ChangePoint
	Merge(other ChangePoint) (ChangePoint, error)
}

func NewChangePoint(commit string, test TestResult) (ChangePoint, error) {
	if test == nil {
		return nil, fmt.Errorf("Parameter test is nil")
	}

	_, ok := test.ExecutionResults(commit)
	testName := test.Test()
	if !ok {
		return nil, fmt.Errorf("Commit '%s' is not contained in TestResult for test '%s'", commit, test.Test())
	}

	t, err := ChangePointTypeFromResult(commit, test)
	if err != nil {
		return nil, err
	}

	return &cp{
		C:   commit,
		Tns: []string{testName},
		ers: map[string]TestResult{
			testName: test,
		},
		T: t,
	}, nil
}

type cp struct {
	C   string   `json:"Commit"`
	Tns []string `json:"TestNames"`
	ers map[string]TestResult
	l   sync.RWMutex
	T   ChangePointType `json:"Type"`
}

func (c *cp) TestNames() []string {
	c.l.RLock()
	defer c.l.RUnlock()
	return c.Tns
}

func (c *cp) Commit() string {
	c.l.RLock()
	defer c.l.RUnlock()
	return c.C
}

func (c *cp) Type() ChangePointType {
	c.l.RLock()
	defer c.l.RUnlock()
	return c.T
}

func (c *cp) Add(commit string, test TestResult) error {
	if test == nil {
		return fmt.Errorf("Parameter test nil")
	}
	c.l.Lock()
	defer c.l.Unlock()
	if c.C != commit {
		return fmt.Errorf("Invalid commit '%s'. This ChangePoint deals with commit '%s'", commit, c.C)
	}

	// check if changepoints are of same type
	ncp, err := NewChangePoint(commit, test)
	if err != nil {
		return err
	}

	otherType := ncp.Type()
	if c.T != otherType {
		return ChangePointTypeError(fmt.Errorf("ChangePoint.Merge - types are not compatible: %v != %v", c.T, otherType))
	}

	testName := test.Test()
	c.Tns = append(c.Tns, testName)
	c.ers[testName] = test
	return nil
}

func (c *cp) Get(testName string) (TestResult, bool) {
	c.l.RLock()
	defer c.l.RUnlock()
	er, ok := c.ers[testName]
	return er, ok
}

func (c *cp) Merge(other ChangePoint) (ChangePoint, error) {
	if other == nil {
		return c.Copy(), nil
	}
	c.l.RLock()
	defer c.l.RUnlock()

	// check whether change points are of same type
	otherType := other.Type()
	if c.T != otherType {
		return nil, ChangePointTypeError(fmt.Errorf("ChangePoint.Merge - types are not compatible: %v != %v", c.T, otherType))
	}

	oc := other.Commit()
	if c.C != oc {
		return nil, fmt.Errorf("Commits not equal: '%s' != '%s'", c.C, oc)
	}

	// overlapping testnames are not taken into account, hence the underlaying array of tn might be larger than
	otns := other.TestNames()
	tns := make([]string, 0, len(c.Tns)+len(otns))
	m := make(map[string]TestResult)
	// add other change points
	for _, otn := range otns {
		er, ok := other.Get(otn)
		if !ok {
			panic(fmt.Sprintf("Inconsistency between testName (%s) and ChangePoint.Get method", otn))
		}
		m[otn] = er
		tns = append(tns, otn)
	}
	// add c change points
	for k, v := range c.ers {
		// check if other change point already added this ExecutionResult (hence this test in this commit)
		_, ok := m[k]
		if !ok {
			m[k] = v
			tns = append(tns, k)
		} else {
			return nil, fmt.Errorf("!!! Tried to merge two ExecutionResults for test '%s' and commit %s", k, oc)
		}
	}

	return &cp{
		C:   oc,
		ers: m,
		Tns: tns,
		T:   c.T,
	}, nil
}

func (c *cp) Copy() ChangePoint {
	c.l.RLock()
	defer c.l.RUnlock()

	tns := make([]string, len(c.Tns))
	ers := make(map[string]TestResult)
	for i, tn := range c.Tns {
		tns[i] = tn
		ers[tn] = c.ers[tn]
	}

	return &cp{
		C:   c.C,
		Tns: tns,
		ers: ers,
		T:   c.T,
	}
}
