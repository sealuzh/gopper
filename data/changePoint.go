package data

import (
	"fmt"
	"sort"
	"sync"

	"github.com/montanaflynn/stats"
)

const (
	cpsCap          = 10
	cpImprPrefix    = "im_"
	cpRegrPrefix    = "re_"
	jsonImprovement = "improvement"
	jsonRegression  = "regression"
)

// ChangePoints
type ChangePoints interface {
	sort.Interface
	All() []ChangePoint
	Get(commit string, t ChangePointType) (ChangePoint, bool)
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
	if t.IsImprovement() {
		return cpImprPrefix + commit
	}
	return cpRegrPrefix + commit
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

	_, ok := test.ExecutionResult(commit)
	testName := test.Test()
	if !ok {
		return nil, fmt.Errorf("Commit '%s' is not contained in TestResult for test '%s'", commit, test.Test())
	}

	t, err := NewChangePointType(commit, test)
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

// change point type
type ChangePointType interface {
	IsRegression() bool
	IsImprovement() bool
}

func NewChangePointType(commit string, testResult TestResult) (ChangePointType, error) {
	commits := testResult.Commits()
	l := len(commits)
	var commit2 string
	for i, c := range commits {
		if c == commit {
			if i == (l - 1) {
				// last commit in test result
				fmt.Printf("ERROR - ChangePoint as last commit (%s) of test result (%s)\n", c, testResult.Test())
				// should never happen
				return RegressionType, nil
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

	if c1Mean < c2Mean {
		return ImprovementType, nil
	}
	return RegressionType, nil
}

type cpt int

const (
	ImprovementType cpt = iota
	RegressionType
)

func (t cpt) IsRegression() bool {
	return t == RegressionType
}

func (t cpt) IsImprovement() bool {
	return t == ImprovementType
}

func (t cpt) String() string {
	switch t {
	case ImprovementType:
		return jsonImprovement
	case RegressionType:
		return jsonImprovement
	default:
		return fmt.Sprintf("Unknown ChangePointType: %d", t)
	}
}

func (t cpt) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

type ChangePointTypeError error
