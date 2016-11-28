package data

import (
	"fmt"
	"sort"
	"sync"
)

const (
	cpsCap = 10
)

// ChangePoints
type ChangePoints interface {
	sort.Interface
	All() []ChangePoint
	Get(commit string) (ChangePoint, bool)
	Copy() ChangePoints
	Add(c ChangePoint) error
}

func NewChangePoints() ChangePoints {
	return &cps{
		cps:     make(map[string]ChangePoint),
		Commits: make([]ChangePoint, 0, cpsCap),
	}
}

type cps struct {
	l       sync.RWMutex
	cps     map[string]ChangePoint
	Commits []ChangePoint
}

func (c *cps) checkPostConditions() {
	lc := len(c.Commits)
	lcps := len(c.cps)
	if lc != lcps {
		panic(fmt.Sprintf("commits and cps are not of same size: %d != %d", lc, lcps))
	}
}

func (c *cps) All() []ChangePoint {
	c.l.RLock()
	defer c.l.RUnlock()
	cps := make([]ChangePoint, len(c.Commits))
	copy(cps, c.Commits)
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
		for i, oldCp := range c.Commits {
			if oldCp.Commit() == cpc {
				c.Commits[i] = mergedCp
				break
			}
		}
	} else {
		c.cps[cpc] = cp
		c.Commits = append(c.Commits, cp)
	}
	c.checkPostConditions()
	return nil
}

func (c *cps) Copy() ChangePoints {
	c.l.RLock()
	defer c.l.RUnlock()
	cs := make([]ChangePoint, len(c.Commits))
	m := make(map[string]ChangePoint)
	for i, cp := range c.Commits {
		copy := cp.Copy()
		cs[i] = copy
		m[copy.Commit()] = copy
	}
	copy := &cps{
		Commits: cs,
		cps:     m,
	}

	// check if it is a copy
	func() {
		loc := len(c.Commits)
		locps := len(c.cps)
		lold := loc + locps
		lnc := len(copy.Commits)
		lncps := len(copy.cps)
		lnew := lnc + lncps
		if lold != lnew {
			panic(fmt.Sprintf("copy not correct:\nold commits=%d; old cps:%d\nnew commits=%d; new cps=%d", loc, locps, lnc, lncps))
		}
	}()

	return copy
}

func (c *cps) Len() int {
	c.l.RLock()
	defer c.l.RUnlock()
	return len(c.Commits)
}

func (c *cps) Less(i, j int) bool {
	c.l.RLock()
	defer c.l.RUnlock()
	cpi := c.Commits[i]
	cpj := c.Commits[j]
	return len(cpi.TestNames()) < len(cpj.TestNames())
}

func (c *cps) Swap(i, j int) {
	c.l.Lock()
	defer c.l.Unlock()
	buf := c.Commits[i]
	c.Commits[i] = c.Commits[j]
	c.Commits[j] = buf
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
		C:   er.SHA,
		Tns: []string{er.Test},
		ers: map[string]*ExecutionResult{
			er.Test: er,
		},
	}, nil
}

type cp struct {
	C   string
	Tns []string
	ers map[string]*ExecutionResult
	l   sync.RWMutex
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

func (c *cp) Add(er *ExecutionResult) error {
	if er == nil {
		return fmt.Errorf("Parameter is nil")
	}
	c.l.Lock()
	defer c.l.Unlock()
	if c.C != er.SHA {
		return fmt.Errorf("Invalid commit '%s'. This ChangePoint deals with commit '%s'", er.SHA, c.C)
	}
	c.Tns = append(c.Tns, er.Test)
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
	if c.C != oc {
		return nil, fmt.Errorf("Commits not equal: '%s' != '%s'", c.C, oc)
	}

	// overlapping testnames are not taken into account, hence the underlaying array of tn might be larger than
	otns := other.TestNames()
	tns := make([]string, 0, len(c.Tns)+len(otns))
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
		C:   oc,
		ers: m,
		Tns: tns,
	}, nil
}

func (c *cp) Copy() ChangePoint {
	c.l.RLock()
	defer c.l.RUnlock()

	tns := make([]string, len(c.Tns))
	ers := make(map[string]*ExecutionResult)
	for i, tn := range c.Tns {
		tns[i] = tn
		ers[tn] = c.ers[tn]
	}

	return &cp{
		C:   c.C,
		Tns: tns,
		ers: ers,
	}

}
