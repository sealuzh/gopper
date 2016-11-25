package data

import (
	"fmt"
	"sync"
)

const (
	cpsCap = 10
)

// ChangePoints
type ChangePoints interface {
	All() []ChangePoint
	Get(commit string) (ChangePoint, bool)
	Copy() ChangePoints
	Add(c ChangePoint) error
}

func NewChangePoints() ChangePoints {
	return &cps{
		cps:     make(map[string]ChangePoint),
		commits: make([]ChangePoint, 0, cpsCap),
	}
}

type cps struct {
	l       sync.RWMutex
	cps     map[string]ChangePoint
	commits []ChangePoint
}

func (c *cps) All() []ChangePoint {
	c.l.RLock()
	defer c.l.RUnlock()
	cps := make([]ChangePoint, len(c.commits))
	copy(cps, c.commits)
	return cps
}

func (c *cps) Get(commit string) (ChangePoint, bool) {
	c.l.RLock()
	defer c.l.RUnlock()
	cp, ok := c.cps[commit]
	return cp, ok
}

func (c *cps) Add(cp ChangePoint) error {
	if cp == nil {
		return fmt.Errorf("Parameter is nil")
	}
	c.l.Lock()
	defer c.l.Unlock()
	cpc := cp.Commit()
	e, ok := c.cps[cpc]
	if ok {
		mergedCp, err := e.Merge(cp)
		if err != nil {
			return err
		}
		c.cps[cpc] = mergedCp

		// replace change point in commits with new merged change point
		for i, oldCp := range c.commits {
			if oldCp.Commit() == cpc {
				c.commits[i] = mergedCp
				break
			}
		}
	} else {
		c.cps[cpc] = cp
		c.commits = append(c.commits, cp)
	}
	return nil
}

func (c *cps) Copy() ChangePoints {
	cs := make([]ChangePoint, len(c.commits))
	m := make(map[string]ChangePoint)
	for _, cp := range c.commits {
		copy := cp.Copy()
		cs = append(cs, copy)
		m[copy.Commit()] = copy
	}
	return &cps{
		commits: cs,
		cps:     m,
	}
}

// ChangePoint
type ChangePoint interface {
	TestNames() []string
	Commit() string
	Add(er *ExecutionResult) error
	Get(testName string) (*ExecutionResult, bool)
	Copy() ChangePoint
	Merge(other ChangePoint) (ChangePoint, error)
}

func NewChangePoint(er *ExecutionResult) (ChangePoint, error) {
	if er == nil {
		return nil, fmt.Errorf("Parameter es is nil")
	}
	return &cp{
		commit:    er.SHA,
		testNames: []string{er.Test},
		ers: map[string]*ExecutionResult{
			er.Test: er,
		},
	}, nil
}

type cp struct {
	commit    string
	testNames []string
	ers       map[string]*ExecutionResult
	l         sync.RWMutex
}

func (c *cp) TestNames() []string {
	c.l.RLock()
	defer c.l.RUnlock()
	return c.testNames
}

func (c *cp) Commit() string {
	c.l.RLock()
	defer c.l.RUnlock()
	return c.commit
}

func (c *cp) Add(er *ExecutionResult) error {
	if er == nil {
		return fmt.Errorf("Parameter is nil")
	}
	c.l.Lock()
	defer c.l.Unlock()
	if c.commit != er.SHA {
		return fmt.Errorf("Invalid commit '%s'. This ChangePoint deals with commit '%s'", er.SHA, c.commit)
	}
	c.testNames = append(c.testNames, er.Test)
	c.ers[er.Test] = er
	return nil
}

func (c *cp) Get(testName string) (*ExecutionResult, bool) {
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
	oc := other.Commit()
	if c.commit != oc {
		return nil, fmt.Errorf("Commits not equal: '%s' != '%s'", c.commit, oc)
	}

	// overlapping testnames are not taken into account, hence the underlaying array of tn might be larger than
	otns := other.TestNames()
	tns := make([]string, 0, len(c.testNames)+len(otns))
	m := make(map[string]*ExecutionResult)
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
		commit:    oc,
		ers:       m,
		testNames: tns,
	}, nil
}

func (c *cp) Copy() ChangePoint {
	c.l.RLock()
	defer c.l.RUnlock()

	tns := make([]string, len(c.testNames))
	ers := make(map[string]*ExecutionResult)
	for i, tn := range c.testNames {
		tns[i] = tn
		ers[tn] = c.ers[tn]
	}

	return &cp{
		commit:    c.commit,
		testNames: tns,
		ers:       ers,
	}

}
