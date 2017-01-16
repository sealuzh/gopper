package data

import (
	"fmt"
	"sort"
	"sync"
)

const (
	cpsCap           = 10
	cpPrefixTemplate = "%s_%d_%s"
	cpImprPrefix     = "im"
	cpRegrPrefix     = "re"
)

// ChangePoints
type ChangePoints interface {
	sort.Interface
	All() []ChangePoint
	Get(commit string, t ChangePointType) (ChangePoint, bool)
	At(commit string) ChangePoints
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

func cpKeyCp(c ChangePoint) string {
	commit := c.Commit()
	return cpKeyCt(commit, c.Type())
}

func cpKeyCt(commit string, t ChangePointType) string {
	cat := t.Category()
	if t.IsImprovement() {
		return fmt.Sprintf(cpPrefixTemplate, cpImprPrefix, cat, commit)
	} else if t.IsRegression() {
		return fmt.Sprintf(cpPrefixTemplate, cpRegrPrefix, cat, commit)
	}
	panic("unknown change point type")
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

func (c *cps) Get(commit string, t ChangePointType) (ChangePoint, bool) {
	c.l.RLock()
	defer c.l.RUnlock()
	cp, ok := c.cps[cpKeyCt(commit, t)]
	return cp, ok
}

func (c *cps) Add(cp ChangePoint) error {
	if cp == nil {
		return fmt.Errorf("Parameter is nil")
	}
	c.l.Lock()
	defer c.l.Unlock()
	keyCp := cpKeyCp(cp)
	e, ok := c.cps[keyCp]
	otherType := cp.Type()
	// if change point with this commit was already added and is of the same type
	if ok && e.Type() == otherType {
		mergedCp, err := e.Merge(cp)
		if err != nil {
			return err
		}
		c.cps[keyCp] = mergedCp

		// replace change point in commits with new merged change point
		for i, oldCp := range c.Commits {
			if oldCp.Commit() == cp.Commit() && oldCp.Type() == otherType {
				c.Commits[i] = mergedCp
				break
			}
		}
	} else {
		c.cps[keyCp] = cp
		c.Commits = append(c.Commits, cp)
	}
	c.checkPostConditions()
	return nil
}

func (c *cps) At(commit string) ChangePoints {
	c.l.RLock()
	defer c.l.RUnlock()
	cps := NewChangePoints()
	for _, t := range AllChangePointTypes() {
		cp, ok := c.cps[cpKeyCt(commit, t)]
		if ok {
			cps.Add(cp)
		}
	}
	return cps
}

func (c *cps) Copy() ChangePoints {
	c.l.RLock()
	defer c.l.RUnlock()
	cs := make([]ChangePoint, len(c.Commits))
	m := make(map[string]ChangePoint)
	for i, cp := range c.Commits {
		copy := cp.Copy()
		cs[i] = copy
		m[cpKeyCp(copy)] = copy
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
