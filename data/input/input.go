package input

type Config struct {
	In        []string
	Out       Out
	Transform []Func
	Analyse   Func
}

type Func struct {
	Name   string
	Params []interface{}
}

type SubPrograms struct {
	Count       int
	List        []string
	Occurrences map[string][]int
}

type Out struct {
	TestResults  []string
	ChangePoints []string
	Plot         string
}
